# Implementation Tasks

## 1. Editor Detection
- [ ] 1.1 Check `$VISUAL` environment variable first
- [ ] 1.2 Fall back to `$EDITOR` if `$VISUAL` not set
- [ ] 1.3 Error if neither is set with helpful message

## 2. Open in Editor
- [ ] 2.1 Implement `OpenInEditor(path string)` function
- [ ] 2.2 Execute editor command with workspace path
- [ ] 2.3 Handle editor exit (wait vs background based on editor type)

## 3. Browser URL Resolution
- [ ] 3.1 Add `GetRemoteURL(repoPath string)` to gitx package
- [ ] 3.2 Parse git remote URL to HTTPS URL
- [ ] 3.3 Handle SSH URLs (git@github.com:user/repo.git -> https://github.com/user/repo)
- [ ] 3.4 Handle various hosts (GitHub, GitLab, Bitbucket)

## 4. Open in Browser
- [ ] 4.1 Implement `OpenInBrowser(url string)` using `open` (macOS) / `xdg-open` (Linux)
- [ ] 4.2 Open all repos if no `--repo` specified
- [ ] 4.3 Open specific repo if `--repo` specified

## 5. CLI Command
- [ ] 5.1 Create `workspaceOpenCmd` cobra command
- [ ] 5.2 Add `--browser` flag (default: false, opens editor)
- [ ] 5.3 Add `--repo` flag to target specific repo

## 6. Testing
- [ ] 6.1 Test editor detection priority
- [ ] 6.2 Test SSH to HTTPS URL conversion
- [ ] 6.3 Test browser open on various platforms
