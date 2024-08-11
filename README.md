[![progress-banner](https://backend.codecrafters.io/progress/git/10c8f19f-00b9-4cb1-9faf-0e5480b3838b)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

This is a solution for
["Build Your Own Git" Challenge](https://codecrafters.io/challenges/git).

In this challenge, you'll build a small Git implementation that's capable of
initializing a repository, creating commits and cloning(clone is WIP) a public repository.
Along the way we'll learn about the `.git` directory, Git objects (blobs,
commits, trees etc.), Git's transfer protocols and more.

**Note**: If you're viewing this repo on GitHub, head over to
[codecrafters.io](https://codecrafters.io) to try the challenge.


It allows only the following commands:
### Init
Initializes an empty .git directory in the current folder
```sh
./your_git.sh init
```

### Catfile
Displays the content of an object given it's hash
```sh
./your_git.sh cat-file <object_hash>
```

### Hash-Object
Creates the hash of the given file
```sh
./your_git.sh hash-object <file_name>
```

### Write-Tree/add
Adds the current state of changes to git
```sh
./your_git.sh write-tree
```

### Commit-Tree/Commit
Commits the tree object given tree hash
```sh
./your_git.sh commit-tree <tree_hash>
```

### Ls-Tree
Displays the content of Tree object
```sh
./your_git.sh ls-tree <tree_hash>
```

### Clone
Clones from the given https link(incomplete)
```sh
./your_git.sh clone <repo_link>
```