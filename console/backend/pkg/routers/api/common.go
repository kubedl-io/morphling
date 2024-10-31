package api

import (
	"github.com/alibaba/morphling/console/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
)

func handleErr(c *gin.Context, msg string) {
	formattedMsg := msg
	klog.Error(formattedMsg)
	utils.Failed(c, msg)
}

func pushToGitHub(filePath string, content []byte, gitHubRepoInfo utils.GitHubRepoInfo) error {
	// Setup GitHub client
	ctx := context.Background()

	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gitHubRepoInfo.AccessToken},
	)))

	// Check if file exists
	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		gitHubRepoInfo.Owner,
		gitHubRepoInfo.Repo,
		filePath,
		&github.RepositoryContentGetOptions{Ref: gitHubRepoInfo.Branch},
	)
	if err != nil && !strings.Contains(err.Error(), "404") {
		klog.Errorf("error checking file existence: %v", err)
		return err
	}

	// Prepare commit message
	commitMessage := fmt.Sprintf("Update %s", filePath)
	var sha *string
	if fileContent != nil {
		sha = fileContent.SHA
	}

	// Create or update file
	_, _, err = client.Repositories.CreateFile(ctx, gitHubRepoInfo.Owner, gitHubRepoInfo.Repo, filePath, &github.RepositoryContentFileOptions{
		Message: &commitMessage,
		Content: content,
		SHA:     sha,
		Branch:  &gitHubRepoInfo.Branch,
	})

	if err != nil {
		klog.Errorf("error pushing file to GitHub: %v", err)
		return err
	}

	return nil
}