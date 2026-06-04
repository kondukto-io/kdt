/*
Copyright © 2019 Invicti Security
https://www.invicti.com/
*/

package cmd

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kondukto-io/kdt/client"
)

func TestScan_prepareCustomParams(t *testing.T) {
	type args struct {
		key         string
		custom      *client.Custom
		parsedValue interface{}
	}
	tests := []struct {
		name string
		args args
		want *client.Custom
	}{
		{
			name: "valid - path with one dot",
			args: args{
				key: "ruleset_options.exclude",
				custom: &client.Custom{
					Params: map[string]interface{}{
						"ruleset_options": map[string]interface{}{
							"ruleset": "gosec",
						},
					},
				},
				parsedValue: "vendor",
			},
			want: &client.Custom{
				Params: map[string]interface{}{
					"ruleset_options": map[string]interface{}{
						"ruleset": "gosec",
						"exclude": "vendor",
					},
				},
			},
		},
		{
			name: "valid - path without dot",
			args: args{
				key: "ruleset_type",
				custom: &client.Custom{
					Params: map[string]interface{}{
						"ruleset_options": map[string]interface{}{
							"ruleset": "gosec",
						},
					},
				},
				parsedValue: 0,
			},
			want: &client.Custom{
				Params: map[string]interface{}{
					"ruleset_type": 0,
					"ruleset_options": map[string]interface{}{
						"ruleset": "gosec",
					},
				},
			},
		},
		{
			name: "valid - path with two dot",
			args: args{
				key: "image.image_detail.hash",
				custom: &client.Custom{
					Params: map[string]interface{}{
						"sha": 1234,
						"image": map[string]interface{}{
							"tag": "latest",
						},
					},
				},
				parsedValue: "12345890",
			},
			want: &client.Custom{
				Params: map[string]interface{}{
					"sha": 1234,
					"image": map[string]interface{}{
						"tag": "latest",
						"image_detail": map[string]interface{}{
							"hash": "12345890",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := appendKeyToParamsMap(tt.args.key, tt.args.custom, tt.args.parsedValue)
			if err != nil {
				t.Errorf("Scan.prepareCustomParams() error = %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scan.prepareCustomParams() = %v, want %v", got, tt.want)
			}
			fmt.Printf("%+v\n", got.Params)
		})
	}
}

// TestScan_almToolFlag guards against a regression where the scan command's
// alm-tool flag was registered with String("alm-tool", "A", ...), making "A"
// the default value instead of the -A shorthand. With that bug, an auto-created
// project (--create-project) would be sent to the API with ALM tool "A" when
// the user did not explicitly pass --alm-tool.
func TestScan_almToolFlag(t *testing.T) {
	flag := scanCmd.Flags().Lookup("alm-tool")
	if flag == nil {
		t.Fatal("alm-tool flag is not registered on the scan command")
	}

	if flag.DefValue != "" {
		t.Errorf("alm-tool default value = %q, want empty string", flag.DefValue)
	}

	if flag.Shorthand != "A" {
		t.Errorf("alm-tool shorthand = %q, want %q", flag.Shorthand, "A")
	}

	got, err := scanCmd.Flags().GetString("alm-tool")
	if err != nil {
		t.Fatalf("failed to read alm-tool flag: %v", err)
	}
	if got != "" {
		t.Errorf("alm-tool value with no flag set = %q, want empty string", got)
	}
}
