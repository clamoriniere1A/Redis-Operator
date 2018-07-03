package rediscluster

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
)

func TestGetResourcesRequirements(t *testing.T) {
	type args struct {
		resourceConf ClusterNodeResourcesConf
	}
	tests := []struct {
		name    string
		args    args
		want    v1.ResourceRequirements
		wantErr bool
	}{
		{
			name: "empty resources conf",
			args: args{
				resourceConf: ClusterNodeResourcesConf{},
			},
			want:    v1.ResourceRequirements{},
			wantErr: false,
		},
		/*{
			name: "memory requested set",
			args: args{
				resourceConf: ClusterNodeResourcesConf{
					MemoryRequest: "1Gi",
				},
			},
			want:    v1.ResourceRequirements{},
			wantErr: false,
		},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetResourcesRequirements(tt.args.resourceConf)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResourcesRequirements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResourcesRequirements() = %v, want %v", got, tt.want)
			}
		})
	}
}
