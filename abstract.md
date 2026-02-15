### abstract
this is a very old project that i used as my golang learning exercise. while it provides helpful functionality and works as intended in managing wireguard configurations, it is rather clunky.

### plan
i plan on reviving it with further development. however, before doing so, i want to shape it up and bring to modern standards for writing go code.

### do the following:
- analyze the codebase to understand the functionality. create a docs/FUNC.md file that outlines the core functionalities of the project, including how to use it, its commands, and expected behaviors.
- create a test suite to cover end to end functionality. focus on the end to end, no unit tests. the idea is that after refactoring, the core functionality for adding peers, generating configs, etc still works as intended. the test suite should serve as an *acceptance test* for the project and be structured as such.
- create a new branch named `refactor-modernize` for the refactoring work and make a commit with the test suite added.
- prepare a detailed refactoring plan that outlines the steps needed to modernize the codebase and save it in a docs/REFACTORING-PLAN.md file. this should include:
  - updating dependencies to their latest versions
  - restructuring the project layout to follow current best practices
  - improving code readability and maintainability
  - implementing error handling and logging where necessary
  - adding comprehensive test coverage for core functionalities
- perform the refactoring based on the plan, ensuring that all tests pass after each significant change. the acceptance test suite from step 2 still needs to pass after refactoring.

### notes
- you are free to make any changes within the repository as long as the acceptance tests pass before and after the refactoring and cover all commands available in the project
- the repo will no longer use azure pipelines for ci/cd. you can remove any related files or configurations. the repo will only use github actions from now on. there is a github action already created. it can be updated as needed but should still have the same functionality and deliverables.
- *DO NOT* make any modifications in snapcraft.yaml under any circumstance