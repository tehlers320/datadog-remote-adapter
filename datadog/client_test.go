package datadog

import (
	"reflect"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/prometheus/prometheus/prompb"
)

var (
	BasicDatapoints = [][]*float64{
		{datadog.PtrFloat64(1600000000), datadog.PtrFloat64(1.0)},
		{datadog.PtrFloat64(1600000010), datadog.PtrFloat64(1.0)},
		{datadog.PtrFloat64(1600000020), datadog.PtrFloat64(1.0)},
		{datadog.PtrFloat64(1600000030), datadog.PtrFloat64(9.0)},
		{datadog.PtrFloat64(1600000040), datadog.PtrFloat64(1.0)},
		{datadog.PtrFloat64(1600000050), datadog.PtrFloat64(1.0)},
	}
)

func MakeCommonMetrics() datadogV1.MetricsQueryResponse {
	//kubernetes_state_container_memory_requested
	K8sContainerMemory := datadogV1.MetricsQueryResponse{
		Series: []datadogV1.MetricsQueryMetadata{
			{
				Aggr:        datadog.NullableString{},
				DisplayName: nil,
				End:         nil,
				Expression:  nil,
				Interval:    nil,
				Length:      datadog.PtrInt64(6),
				Metric:      nil,
				Pointlist:   BasicDatapoints,
				QueryIndex:  nil,
				Scope:       nil,
				Start:       nil,
				TagSet:      []string{},
				Unit:        []datadogV1.MetricsQueryUnit{},
				UnparsedObject: map[string]interface{}{
					"": nil,
				},
				AdditionalProperties: map[string]interface{}{
					"": nil,
				},
			},
		},
	}
	return K8sContainerMemory
}

func Test_valuesToSamples(t *testing.T) {
	type args struct {
		values [][]*float64
	}
	tests := []struct {
		name    string
		args    args
		want    []prompb.Sample
		wantErr bool
	}{
		{
			name: "",
			args: args{
				values: BasicDatapoints,
			},
			want: []prompb.Sample{
				{
					Value:     1.0,
					Timestamp: 1600000000,
				},
				{
					Value:     1.0,
					Timestamp: 1600000010,
				},
				{
					Value:     1.0,
					Timestamp: 1600000020,
				},
				{
					Value:     9.0,
					Timestamp: 1600000030,
				},
				{
					Value:     1.0,
					Timestamp: 1600000040,
				},
				{
					Value:     1.0,
					Timestamp: 1600000050,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := valuesToSamples(tt.args.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("valuesToSamples() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("valuesToSamples() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mergeSamples(t *testing.T) {
	type args struct {
		a []prompb.Sample
		b []prompb.Sample
	}
	tests := []struct {
		name string
		args args
		want []prompb.Sample
	}{
		{
			name: "",
			args: args{
				a: []prompb.Sample{
					{
						Value:     1.0,
						Timestamp: 1600000000,
					},
					{
						Value:     1.0,
						Timestamp: 1600000010,
					},
					{
						Value:     1.0,
						Timestamp: 1600000020,
					},
					{
						Value:     9.0,
						Timestamp: 1600000030,
					},
					{
						Value:     1.0,
						Timestamp: 1600000040,
					},
					{
						Value:     1.0,
						Timestamp: 1600000050,
					},
				},
				b: []prompb.Sample{
					{
						Value:     2.0,
						Timestamp: 1600000000,
					},
					{
						Value:     2.0,
						Timestamp: 1600000010,
					},
					{
						Value:     2.0,
						Timestamp: 1600000020,
					},
					{
						Value:     9.0,
						Timestamp: 1600000030,
					},
					{
						Value:     2.0,
						Timestamp: 1600000040,
					},
					{
						Value:     2.0,
						Timestamp: 1600000050,
					},
				},
			},
			want: []prompb.Sample{
				{
					Value:     1.0,
					Timestamp: 1600000000,
				},
				{
					Value:     1.0,
					Timestamp: 1600000010,
				},
				{
					Value:     1.0,
					Timestamp: 1600000020,
				},
				{
					Value:     9.0,
					Timestamp: 1600000030,
				},
				{
					Value:     1.0,
					Timestamp: 1600000040,
				},
				{
					Value:     1.0,
					Timestamp: 1600000050,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeSamples(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeSamples() = %v, want %v", got, tt.want)
			}
		})
	}
}
