package label_test

import (
	"strconv"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/fake"
	"github.com/jenkins-x/jx-gitops/pkg/cmd/pr/label"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner/fakerunner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullRequestLabel(t *testing.T) {
	testCases := []struct {
		name        string
		init        func(o *label.Options, pr *scm.PullRequest, fakeData *fake.Data)
		verify      func(o *label.Options, pr *scm.PullRequest)
		expectError bool
	}{
		{
			name: "already has label",
			init: func(o *label.Options, pr *scm.PullRequest, fakeData *fake.Data) {
				o.Label = "mylabel"
				fakeData.IssueLabelsExisting = []string{AddPullRequestLabel(pr, "mylabel")}
			},
			verify: func(o *label.Options, pr *scm.PullRequest) {
				assert.False(t, o.LabelAdded)
			},
		},
		{
			name: "add label",
			init: func(o *label.Options, pr *scm.PullRequest, fakeData *fake.Data) {
				o.Label = "mylabel"
			},
			verify: func(o *label.Options, pr *scm.PullRequest) {
				assert.True(t, o.LabelAdded)
			},
		},
		{
			name: "add label if matching",
			init: func(o *label.Options, pr *scm.PullRequest, fakeData *fake.Data) {
				o.Label = "mylabel"
				o.Regex = "env/.*"
				fakeData.IssueLabelsExisting = []string{AddPullRequestLabel(pr, "env/staging")}
			},
			verify: func(o *label.Options, pr *scm.PullRequest) {
				assert.True(t, o.LabelAdded)
			},
		},
		{
			name: "not add label as not matching",
			init: func(o *label.Options, pr *scm.PullRequest, fakeData *fake.Data) {
				o.Label = "mylabel"
				o.Regex = "env/.*"
				fakeData.IssueLabelsExisting = []string{AddPullRequestLabel(pr, "somethingElse")}
			},
			verify: func(o *label.Options, pr *scm.PullRequest) {
				assert.False(t, o.LabelAdded)
			},
		},
		{
			name:        "missing label name",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		name := tc.name
		t.Logf("test %s:\n", name)

		_, o := label.NewCmdPullRequestLabel()

		prNumber := 123
		repo := "myorg/myrepo"
		prBranch := "my-pr-branch-name"

		scmClient, fakeData := fake.NewDefault()
		o.ScmClient = scmClient
		runner := &fakerunner.FakeRunner{}
		o.CommandRunner = runner.Run
		o.SourceURL = "https://github.com/" + repo
		o.Number = prNumber
		o.Branch = prBranch

		pr := &scm.PullRequest{
			Number: prNumber,
			Title:  "my awesome pull request",
			Body:   "some text",
			Source: prBranch,
			Base: scm.PullRequestBranch{
				Repo: scm.Repository{
					FullName: repo,
				},
			},
			Link: o.SourceURL + "/pull/" + strconv.Itoa(prNumber),
		}
		fakeData.PullRequests[prNumber] = pr

		if tc.init != nil {
			tc.init(o, pr, fakeData)
		}

		err := o.Run()

		if tc.expectError {
			require.Error(t, err, "expected error for test %s", name)
			t.Logf("test %s: got expected error %s\n", name, err.Error())
		} else {
			require.NoError(t, err, "failed to run test %s", name)
		}

		if tc.verify != nil {
			tc.verify(o, pr)
		}
	}
}

func AddPullRequestLabel(pr *scm.PullRequest, labelName string) string {
	repo := pr.Repository()
	label := repo.FullName + "#" + strconv.Itoa(pr.Number) + ":" + labelName
	return label
}
