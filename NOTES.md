# Notes to self #

## Setup ##
- How cobra project was setup:
    - `go mod init exectester`
    - `cobra-cli init`
    - Wrapped `&cobra.Command{}` with func so it can be tested

## Golang Newbie Notes ##
- dump obj
    - `fmt.Printf("%+v\n", someObj)`
- use -v to stream test output
    - `go test -v -timeout 5s -run ....`
- clear the test cache
    - `go clean -testcache`

## Misc ##
- create a unix socket and send text to it
    - `nc -lkU /tmp/bentest.sock`
    - `nc -U /tmp/bentest.sock`
        - type text and ENTER

## TODO List  ##
- everything should probably be private
    - struct fields should be private too
    - Review package level vars
- Review my use of pointers
- Fix exitcode test

- Setup github actions
    - https://github.com/apps/sonarcloud
    - https://docs.github.com/en/actions/security-guides/using-secrets-in-github-actions
- Setup sonar coverage reports
    - Official sonar scanner image: https://hub.docker.com/r/sonarsource/sonar-scanner-cli/tags
- Publish to pkgsite
    - Add a readme that works with pkgsite


- Make TestSocket a separate package
- Maybe add super verbose logging that returns hostname and bunch of other stuff?
    - An excuse to add fields to the log output
- Maybe hack on sphinx + godoc
- Support interrogating and sending output to named pipes