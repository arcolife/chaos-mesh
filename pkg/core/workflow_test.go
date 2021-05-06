// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func Test_convertWorkflow(t *testing.T) {
	type args struct {
		kubeWorkflow v1alpha1.Workflow
	}
	tests := []struct {
		name string
		args args
		want Workflow
	}{
		{
			name: "simple workflow",
			args: args{
				v1alpha1.Workflow{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "fake-workflow-0",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "an-entry",
					},
					Status: v1alpha1.WorkflowStatus{},
				},
			},
			want: Workflow{
				Namespace: "fake-namespace",
				Name:      "fake-workflow-0",
				Entry:     "an-entry",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertWorkflow(tt.args.kubeWorkflow); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertWorkflow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertWorkflowDetail(t *testing.T) {
	type args struct {
		kubeWorkflow v1alpha1.Workflow
		kubeNodes    []v1alpha1.WorkflowNode
	}
	tests := []struct {
		name    string
		args    args
		want    WorkflowDetail
		wantErr bool
	}{
		{
			name: "simple workflow detail with no nodes",
			args: args{
				kubeWorkflow: v1alpha1.Workflow{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "another-namespace",
						Name:      "another-fake-workflow",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry:     "another-entry",
						Templates: nil,
					},
					Status: v1alpha1.WorkflowStatus{},
				},
				kubeNodes: nil,
			},
			want: WorkflowDetail{
				Workflow: Workflow{
					Namespace: "another-namespace",
					Name:      "another-fake-workflow",
					Entry:     "another-entry",
				},
				Topology: Topology{
					Nodes: []Node{},
				},
				KubeObject: KubeObjectDesc{
					Meta: KubeObjectMeta{
						Name:      "another-fake-workflow",
						Namespace: "another-namespace",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "another-entry",
					},
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertWorkflowDetail(tt.args.kubeWorkflow, tt.args.kubeNodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertWorkflowDetail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertWorkflowDetail() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertWorkflowNode(t *testing.T) {
	type args struct {
		kubeWorkflowNode v1alpha1.WorkflowNode
	}
	tests := []struct {
		name    string
		args    args
		want    Node
		wantErr bool
	}{
		{
			name: "simple node",
			args: args{kubeWorkflowNode: v1alpha1.WorkflowNode{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "fake-namespace",
					Name:      "fake-node-0",
				},
				Spec: v1alpha1.WorkflowNodeSpec{
					WorkflowName: "fake-workflow-0",
					TemplateName: "fake-template-0",
					Type:         v1alpha1.TypeJVMChaos,
				},
				Status: v1alpha1.WorkflowNodeStatus{},
			}},
			want: Node{
				Name:     "fake-node-0",
				Type:     ChaosNode,
				Serial:   nil,
				Parallel: nil,
				Template: "fake-template-0",
				State:    NodeRunning,
			},
		}, {
			name: "serial node",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "fake-serial-node-0",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "fake-serial-node",
						WorkflowName: "fake-workflow-0",
						Type:         v1alpha1.TypeSerial,
						Tasks:        []string{"child-0", "child-1"},
					},
					Status: v1alpha1.WorkflowNodeStatus{},
				},
			},
			want: Node{
				Name: "fake-serial-node-0",
				Type: SerialNode,
				Serial: &NodeSerial{
					Tasks: []string{"child-0", "child-1"},
				},
				Parallel: nil,
				Template: "fake-serial-node",
				State:    NodeRunning,
			},
		},
		{
			name: "parallel node",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "parallel-node-0",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "parallel-node",
						WorkflowName: "another-fake-workflow",
						Type:         v1alpha1.TypeParallel,
						Tasks:        []string{"child-1", "child-0"},
					},
					Status: v1alpha1.WorkflowNodeStatus{},
				},
			},
			want: Node{
				Name:   "parallel-node-0",
				Type:   ParallelNode,
				Serial: nil,
				Parallel: &NodeParallel{
					Tasks: []string{"child-1", "child-0"},
				},
				Template: "parallel-node",
				State:    NodeRunning,
			},
		},
		{
			name: "some chaos",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "io-chaos-0",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "io-chaos",
						WorkflowName: "another-workflow-0",
						Type:         v1alpha1.TypeIoChaos,
						EmbedChaos: &v1alpha1.EmbedChaos{
							IoChaos: &v1alpha1.IoChaosSpec{
								Mode:       v1alpha1.OnePodMode,
								Action:     "delay",
								Delay:      "100ms",
								Path:       "/fake/path",
								Percent:    100,
								VolumePath: "/fake/path",
							},
						},
					},
					Status: v1alpha1.WorkflowNodeStatus{},
				},
			},
			want: Node{
				Name:     "io-chaos-0",
				Type:     ChaosNode,
				Serial:   nil,
				Parallel: nil,
				Template: "io-chaos",
				State:    NodeRunning,
			},
		},
		{
			name: "accomplished node",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "the-entry-0",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "the-entry",
						WorkflowName: "fake-workflow-0",
						Type:         v1alpha1.TypeSerial,
						Tasks:        []string{"unimportant-task-0"},
					},
					Status: v1alpha1.WorkflowNodeStatus{
						Conditions: []v1alpha1.WorkflowNodeCondition{
							{
								Type:   v1alpha1.ConditionAccomplished,
								Status: corev1.ConditionTrue,
								Reason: "unit test mocked true",
							},
						},
					},
				},
			},
			want: Node{
				Name:  "the-entry-0",
				Type:  SerialNode,
				State: NodeSucceed,
				Serial: &NodeSerial{
					Tasks: []string{"unimportant-task-0"},
				},
				Parallel: nil,
				Template: "the-entry",
			},
		},
		{
			name: "deadline exceed node",
			args: args{kubeWorkflowNode: v1alpha1.WorkflowNode{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "fake-namespace",
					Name:      "deadline-exceed-node-0",
				},
				Spec: v1alpha1.WorkflowNodeSpec{
					TemplateName: "deadline-exceed-node",
					WorkflowName: "some-workflow",
					Type:         v1alpha1.TypePodChaos,
				},
				Status: v1alpha1.WorkflowNodeStatus{
					Conditions: []v1alpha1.WorkflowNodeCondition{
						{
							Type:   v1alpha1.ConditionDeadlineExceed,
							Status: corev1.ConditionTrue,
							Reason: "unit test mocked true",
						},
					},
				},
			}},
			want: Node{
				Name:     "deadline-exceed-node-0",
				Type:     ChaosNode,
				State:    NodeSucceed,
				Serial:   nil,
				Parallel: nil,
				Template: "deadline-exceed-node",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertWorkflowNode(tt.args.kubeWorkflowNode)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertWorkflowNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertWorkflowNode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
