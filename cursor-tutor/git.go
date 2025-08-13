
// gitCommand is a helper function that executes a git command with the given arguments
func gitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Init is a function that calls external git command to initialize a new git repository
func Init() error {
	return gitCommand("init")
}

// Push is a function that calls external git command to push local commits to remote
func Push() error {
	return gitCommand("push")
}

// Commit is a function that calls external git command to commit changes
func Commit(message string) error {
	return gitCommand("commit", "-m", message)
}

// Add is a function that calls external git command to add updated files to the staging area
func Add() error {
	return gitCommand("add", ".")
}

// CreateNewBranch is a function that calls external git command to create a new branch
func CreateNewBranch(branchName string) error {
	return gitCommand("checkout", "-b", branchName)
}

// ChangeBranch is a function that calls external git command to switch to a different branch
func ChangeBranch(branchName string) error {
	return gitCommand("checkout", branchName)
}
