# Coding Practices

This ia an attempt to describe what my development practices are. This is a living document and will be updated as my practices evolve.

This will be used to inform agents about how I like to code and what practices I follow.

## Git Hygiene

- When I start a new feature, I create a new branch from the main branch.
- After a feature is complete, I squash all my commits into one commit with a descriptive message.
- After the feature is committed, I push the files to the remote repository.
- I create a pull request to merge the feature branch into the main branch.
- I ensure that the code passes all tests and reviews before merging.
- I ensure that the main branch is always deployable.
- I regularly pull changes from the main branch into my feature branches to keep them up to date.
- I use conventional commit messages to maintain a consistent commit history.
- I avoid committing large files or sensitive information to the repository.
- I use `.gitignore` files to exclude unnecessary files from being tracked by Git.
- I write clear and concise commit messages that describe the changes made in each commit.

## Coding Standard
- I prefer to use golang for any development work. I use other languages if a library that would solve my problem is only available in that language.
- I follow the [Google Go Style Guide](https://google.github.io/styleguide/go/) for Go code.
- I use golangci-lint to lint my code and ensure it adheres to the style guide.
- I prefer to use descriptive variable names, and avoid single character variable names.