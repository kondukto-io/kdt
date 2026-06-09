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
	"github.com/spf13/cobra"
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

func TestGetSanitizedFlagStr(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "internal spaces preserved", value: "My Project", want: "My Project"},
		{name: "leading and trailing trimmed", value: "  Padded  ", want: "Padded"},
		{name: "no spaces unchanged", value: "NoSpaces", want: "NoSpaces"},
		{name: "empty string unchanged", value: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("project", "", "")
			if err := cmd.Flags().Set("project", tt.value); err != nil {
				t.Fatalf("failed to set flag: %v", err)
			}
			got, err := getSanitizedFlagStr(cmd, "project")
			if err != nil {
				t.Errorf("getSanitizedFlagStr() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("getSanitizedFlagStr() = %q, want %q", got, tt.want)
			}
		})
	}
}
