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

/*
# Initial config

I want to consolidate things like logger initialization and build/test
config in a single place. While Viper is used for config handling of the
cli app itself the config here needs to be initialized before the Cobra
app itself- so I don't think I can use Viper here.

# Logging config

# TODO IM HERE document

# Build config

This is a little gnarly and kind of an experiment. Using Dagger we can ref
this actual project's code from inside the build scripts. Doing this
excessively would make a mess, but I think its worth leveraging judiciously.
*/
package configs

// TODO IM HERE build out docs for configs package
