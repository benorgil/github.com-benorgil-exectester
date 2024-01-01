// # License
//
/*
Copyright © 2023 Ben Orgil
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// # What is it?
//
// A tool to send arbitrary text to stdout or
// stderr and set the exit code
//
// # Project structure
//
// This is a Cobra app thats lives in the "cmd" package/subfolder: [cmd]
//
// For build automation there is a Dagger pipeline in the
// [github.com/benorgil/exectester/build_pipeline] folder.
//
// # Usage
//
// For the actual cli tool usage just use the help menu, it includes a
// bunch of full examples:
//	et --help
//
// This is a console app not a package so having "examples"
// ([cmd.PkgExamples]) for the internal functions is pretty
// unnecessary, but...
//	- I'm looking for an excuse to get familiar with GoDoc!
//	- I'm using/abusing the example snippets that shell out to call the
//    compiled app to double for "integration" tests
//	- Might still be useful to have playground examples for calling the
//    root command?
//	- I created this useless list to try the new GoDoc list support!
//
package main
