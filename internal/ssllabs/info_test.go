// Copyright 2020 Anas Ait Said Oubrahim

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ssllabs

import "testing"

func TestInfo(t *testing.T) {
	info, err := Info()
	if err != nil {
		t.Errorf("\nerror fetching SSLLabs API info: %v", err)
	}

	// the live API values change over time, so only check that they are populated
	if info.EngineVersion == "" || info.CriteriaVersion == "" || info.MaxAssessments == 0 {
		t.Errorf("\nunexpected empty fields in returned info: %v", info)
	}
}
