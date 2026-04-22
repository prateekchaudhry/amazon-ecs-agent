//go:build unit
// +build unit

// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package dockerapi

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	apierrors "github.com/aws/amazon-ecs-agent/ecs-agent/api/errors"
	"github.com/stretchr/testify/assert"
)

func TestRetriableErrorReturnsFalseForNoSuchContainer(t *testing.T) {
	err := CannotStopContainerError{NoSuchContainerError{}}
	assert.False(t, err.IsRetriableError(), "No such container error should be treated as unretriable docker error")
}

func TestRetriableErrorReturnsTrue(t *testing.T) {
	err := CannotStopContainerError{errors.New("error")}
	assert.True(t, err.IsRetriableError(), "Non unretriable error treated as unretriable docker error")
}

func TestRedactEcrUrls(t *testing.T) {
	testCases := []struct {
		overrideStr string
		err         error
		expectedErr error
	}{
		// NO REDACT CASES
		{
			overrideStr: "111111111111.dkr.ecr.us-west-2.amazonaws.com/myapp",
			err:         fmt.Errorf("Error response from daemon: pull access denied for 111111111111.dkr.ecr.us-west-2.amazonaws.com/myapp, repository does not exist or may require 'docker login': denied: User: arn:aws:sts::111111111111:assumed-role/ecsInstanceRole/i-01a01a01a01a01a01a is not authorized to perform: ecr:BatchGetImage on resource: arn:aws:ecr:us-west-2:111111111111:repository/myapp because no resource-based policy allows the ecr:BatchGetImage action"),
			expectedErr: fmt.Errorf("Error response from daemon: pull access denied for 111111111111.dkr.ecr.us-west-2.amazonaws.com/myapp, repository does not exist or may require 'docker login': denied: User: arn:aws:sts::111111111111:assumed-role/ecsInstanceRole/i-01a01a01a01a01a01a is not authorized to perform: ecr:BatchGetImage on resource: arn:aws:ecr:us-west-2:111111111111:repository/myapp because no resource-based policy allows the ecr:BatchGetImage action"),
		},
		{
			overrideStr: "busybox:latest",
			err:         fmt.Errorf("invalid reference format"),
			expectedErr: fmt.Errorf("invalid reference format"),
		},
		{
			overrideStr: "busybox:latest",
			err:         fmt.Errorf(""),
			expectedErr: fmt.Errorf(""),
		},
		{
			overrideStr: "busybox:latest",
			err:         fmt.Errorf("busybox:latest"),
			expectedErr: fmt.Errorf("busybox:latest"),
		},
		{
			overrideStr: "ubuntu",
			err:         fmt.Errorf("busybox:latest"),
			expectedErr: fmt.Errorf("busybox:latest"),
		},
		{
			overrideStr: "111111111111.dkr.ecr.us-west-2.amazonaws.com/myapp",
			err:         fmt.Errorf("String with https://111111111111.dkr.ecr.us-west-2.amazonaws.com/myapp"),
			expectedErr: fmt.Errorf("String with https://111111111111.dkr.ecr.us-west-2.amazonaws.com/myapp"),
		},
		// REDACT CASES
		{
			overrideStr: "111111111111.dkr.ecr.us-east-1.amazonaws.com/myapp",
			err:         fmt.Errorf("ref pull has been retried 5 time(s): failed to copy: httpReadSeeker: failed open: failed to do request: Get \"https://prod-us-east-1-starport-layer-bucket.s3.us-east-1.amazonaws.com/d1d1d1-222222222222-11d11d1d-4444-aaaa-bbbb-e4444gg4g4g4/9s9s9s9s9s-2b2b-a2a2-dddd-gg99gg99gg99?X-Amz-Security-Token=mysup3rs3cr3ttok3n\""),
			expectedErr: fmt.Errorf("ref pull has been retried 5 time(s): failed to copy: httpReadSeeker: failed open: failed to do request: Get REDACTED ECR URL related to 111111111111.dkr.ecr.us-east-1.amazonaws.com/myapp"),
		},
		{
			overrideStr: "111111111111.dkr.ecr.us-east-1.amazonaws.com/myapp",
			err:         fmt.Errorf("https://prod-us-east-1-starport-layer-bucket.s3.us-east-1.amazonaws.com/d1d1d1-222222222222-11d11d1d-4444-aaaa-bbbb-e4444gg4g4g4/9s9s9s9s9s-2b2b-a2a2-dddd-gg99gg99gg99?X-Amz-Security-Token=mysup3rs3cr3tt0k3n"),
			expectedErr: fmt.Errorf("REDACTED ECR URL related to 111111111111.dkr.ecr.us-east-1.amazonaws.com/myapp"),
		},
		{
			overrideStr: "111111111111.dkr.ecr.us-east-1.amazonaws.com/myapp",
			err:         fmt.Errorf("failed to do request: Get \"https://prod-us-east-1-starport-layer-bucket.s3.us-east-1.amazonaws.com/d1d1d1-222222222222-11d11d1d-4444-aaaa-bbbb-e4444gg4g4g4/9s9s9s9s9s-2b2b-a2a2-dddd-gg99gg99gg99?X-Amz-Security-Token=mysup3rs3cr3ttok3n\" and another request \"https://prod-us-east-1-starport-layer-bucket.s3.us-east-1.amazonaws.com/d1d1d1-222222222222-11d11d1d-4444-aaaa-bbbb-e4444gg4g4g4/9s9s9s9s9s-2b2b-a2a2-dddd-gg99gg99gg99?X-Amz-Security-Token=mysup3rs3cr3ttok3n\""),
			expectedErr: fmt.Errorf("failed to do request: Get REDACTED ECR URL related to 111111111111.dkr.ecr.us-east-1.amazonaws.com/myapp and another request REDACTED ECR URL related to 111111111111.dkr.ecr.us-east-1.amazonaws.com/myapp"),
		},
	}

	for _, tc := range testCases {
		redactedErr := redactEcrUrls(tc.overrideStr, tc.err)
		assert.Equal(t, redactedErr.Error(), tc.expectedErr.Error(), "ECR URL redaction output mismatch")
	}
}

func TestImagePlatformMismatchError_Error(t *testing.T) {
	err := ImagePlatformMismatchError{
		Image:     "123456789.dkr.ecr.us-east-1.amazonaws.com/myapp:latest",
		ImageOs:   "linux",
		ImageArch: "arm64",
		HostOs:    "linux",
		HostArch:  "amd64",
	}
	msg := err.Error()
	assert.True(t, strings.Contains(msg, "123456789.dkr.ecr.us-east-1.amazonaws.com/myapp:latest"))
	assert.True(t, strings.Contains(msg, "linux/arm64"))
	assert.True(t, strings.Contains(msg, "linux/amd64"))
	assert.True(t, strings.Contains(msg, "Build the image for the correct platform or use a multi-platform image"))
}

func TestImagePlatformMismatchError_ErrorName(t *testing.T) {
	err := ImagePlatformMismatchError{}
	assert.Equal(t, "ImagePlatformMismatchError", err.ErrorName())
}

func TestImagePlatformMismatchError_Retry(t *testing.T) {
	err := ImagePlatformMismatchError{}
	assert.False(t, err.Retry())
}

func TestImagePlatformMismatchError_ImplementsNamedError(t *testing.T) {
	var _ apierrors.NamedError = ImagePlatformMismatchError{}
}
