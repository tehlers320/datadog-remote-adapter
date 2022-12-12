// Copyright 2022 tehlers
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datadog

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/spf13/viper"
	"golang.org/x/net/context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

// Client allows reading from datadog
type Client struct {
	logger 		log.Logger
	context 	context.Context
	client      *datadogV1.MetricsApi
}

// NewClient creates a new Client.
func NewClient(logger log.Logger) *Client {
	// The default automagicaly gets OS env var DD_API_KEY, DD_APP_KEY and DD_SITE
	ctx := datadog.NewDefaultContext(context.Background())
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV1.NewMetricsApi(apiClient)

	if logger == nil {
		logger = log.NewNopLogger()
	}

	return &Client{
		logger:          logger,
		context:         ctx,
		client:          api,
	}
}

// Read Gets a batch of samples to datadog.
func (c *Client) Read(req *prompb.ReadRequest) (*prompb.ReadResponse, error) {
	labelsToSeries := map[string]*prompb.TimeSeries{}
	for _, q := range req.Queries {
		level.Info(c.logger).Log("msg", "query`: %v\n", q)
		nq, to, from, _ := c.buildQuery(q)
		level.Info(c.logger).Log("msg", "new query`: %v\n", nq)
		resp := c.runQuery(nq, to, from)

		if err := mergeResult(labelsToSeries, []datadogV1.MetricsQueryResponse{resp}); err != nil {
			return nil, err
		}


	}


	resp := prompb.ReadResponse{
		Results: []*prompb.QueryResult{
			{Timeseries: make([]*prompb.TimeSeries, 0, len(labelsToSeries))},
		},
	}
	for _, ts := range labelsToSeries {
		resp.Results[0].Timeseries = append(resp.Results[0].Timeseries, ts)
	}
	return &resp, nil

	
}

func (c *Client)runQuery(query string, to int64, from int64)  datadogV1.MetricsQueryResponse {

	resp, r, err := c.client.QueryMetrics(c.context, from, to, query)

	if err != nil {
		level.Debug(c.logger).Log("msg", "Error when calling `MetricsApi.QueryMetrics`: %v\n", err)
		level.Debug(c.logger).Log("msg", "Full HTTP response: %v\n", r)
	}

	responseContent, _ := json.MarshalIndent(resp, "", "  ")
	level.Debug(c.logger).Log("msg","Response from `MetricsApi.QueryMetrics`:\n%s\n", responseContent)
	return resp
}

// Name identifies the client as a Datadog client.
func (c Client) Name() string {
	return "datadog"
}



func mergeResult(labelsToSeries map[string]*prompb.TimeSeries, results []datadogV1.MetricsQueryResponse) error {
	for _, r := range results {
		for _, s := range r.Series {
			ts, ok := labelsToSeries[*s.DisplayName]
			if !ok {
				ts = &prompb.TimeSeries{
					Labels: nil,
				}
				labelsToSeries[*s.DisplayName] = ts
			}

			samples, err := valuesToSamples(s.Pointlist)
			if err != nil {
				return err
			}

			ts.Samples = mergeSamples(ts.Samples, samples)
		}
	}
	return nil
}


func valuesToSamples(values [][]*float64) ([]prompb.Sample, error) {
	samples := make([]prompb.Sample, 0, len(values))
	for _, v := range values {
		samples = append(samples, prompb.Sample{
			Timestamp: int64(*v[0]),
			Value:     *v[1],
		})
	}
	return samples, nil
}

// mergeSamples merges two lists of sample pairs and removes duplicate
// timestamps. It assumes that both lists are sorted by timestamp.
func mergeSamples(a, b []prompb.Sample) []prompb.Sample {
	result := make([]prompb.Sample, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i].Timestamp < b[j].Timestamp {
			result = append(result, a[i])
			i++
		} else if a[i].Timestamp > b[j].Timestamp {
			result = append(result, b[j])
			j++
		} else {
			result = append(result, a[i])
			i++
			j++
		}
	}
	result = append(result, a[i:]...)
	result = append(result, b[j:]...)
	return result
}


func (c *Client) buildQuery(q *prompb.Query) (string, int64, int64, error) {
	var ddogFormat string
	matchers := make([]string, 0, len(q.Matchers))
	//matchers := make([]string, 0, len(q.Matchers))
	// convert to ddog format 
	// example: system.cpu.idle{host:foo,cluster:bar}} == system_cpu_idle{host="foo",cluster="bar"}
	for _, m := range q.Matchers {
		if m.Name == model.MetricNameLabel {
			switch m.Type {
			case prompb.LabelMatcher_EQ:
				ddogFormat = replaceDot(m.Value)
			default:
				return "", 0, 0, errors.New("name metric does not match anything")
			}
			continue
		}

		switch m.Type {
		case prompb.LabelMatcher_EQ:
			matchers = append(matchers, fmt.Sprintf("%s:'%s'", m.Name, m.Value))
		case prompb.LabelMatcher_NEQ:
			matchers = append(matchers, fmt.Sprintf("!%s:'%s'", m.Name, m.Value))
		case prompb.LabelMatcher_RE:
			matchers = append(matchers, fmt.Sprintf("%s:%s", m.Name, m.Value))
		case prompb.LabelMatcher_NRE:
			matchers = append(matchers, fmt.Sprintf("!%s:%s", m.Name, m.Value))
		default:
			return "", 0, 0, fmt.Errorf("unknown match type %v", m.Type)
		}

	}

	if q.Hints.Func == "sum" {
		ddogFormat = fmt.Sprintf("sum:" + ddogFormat)
	}

	tags := strings.Join(matchers, ",")
	tags = strings.ReplaceAll(tags, "\"", "");
	// Not sure why its adding the single quotes... anyways
	tags = strings.ReplaceAll(tags, "'", "");
	if tags == "" {
		tags = "*"
	}
	ddogFormat = fmt.Sprintf(ddogFormat +  "{" + tags + "}")

	// DDog wont take timestamps in ms....
	// TODO probably check the values are valid. Maybe
	to := toSeconds(q.EndTimestampMs)
	from := toSeconds(q.StartTimestampMs)
	return ddogFormat, to, from, nil


}


func toSeconds(millis int64) int64 {
	return time.Unix(0, millis * int64(time.Millisecond)).Unix()
}

func replaceDot(str string) string { 
	mappings := viper.GetStringMapString("mappings")

	_, ok := mappings[str]
	if ok {
		converted := mappings[str]
		return converted
	} 

	return strings.Replace(str, `_`, `.`, -1)
}