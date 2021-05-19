# Terraform MongoDB Atlas Demo

## Description

This project is designed to create a MongoDB cluster in Atlas, create and
manage some specified users and output the connection strings to connect to
each collection in the MongoDB

## Usage

This code is designed to be run on a POSIX compliant shell with Terraform
installed. It has only been tested in Linux (as that's what I run), but may
run in Mac OS.

Code is called through the "go" script, an interface for both humans and
machines.

### Preparing your controller (workstation)

 1. (Optional) if you use VS Code,
    [a devcontainer has been provided](https://code.visualstudio.com/docs/remote/create-dev-container).
    It is recommended that you run inside the devcontainer as this will provide
    a consistent platform from which to run the code.
 1. Run `./go` to see a help prompt and a list of actions.
 1. Run `./go test controller`, this should report as "Controller environment
    OK" if you're machine is capable of running the code. Remediations will be
    suggested, however in some cases, these can be fixed with
    `./go build controller` and verified by re-running the test. :hand: Note
    that `./go build controller` will just install Terraform and some code
    testing tools into a Python3 virtualenv.
 1. Run `./go test atlas_login` - If you have not set up your environment
    variables, you will need to populate these in your command prompt, eg.

    ```text
    export MONGODB_ATLAS_PROJECT_ID=PROJECT_ID_HERE
    export MONGODB_ATLAS_PUBLIC_KEY=PUBLIC_KEY_HERE
    export MONGODB_ATLAS_PRIVATE_KEY=PRIVATE_KEY_HERE
    ```

    anything that has not been correctly set wil be flagged by the test,
    otherwise a test connection to the MongoDB Atlas will be created and you
    should see "HTTP/2 200" if successful.

### Building the MongoDB cluster

 1. Run `./go build demo` to create the demo environment. You will be prompted
    to confirm that you are happy to build the infrastructure described in
    `terraform plan`.
 1. Once complete, the connections strings for each service will be placed into
    JSON files in the [output/](/output) directory.

### Destroying the MongoDB cluster

 1. Run `./go destroy demo`. Terraform will present you with the actions that
    will be taken and prompt you for a response.

### Cleanup

When you are finished with this code and you want to return the repository to
the state in GitHub, run:

  1. `./go cleanup controller`

## Testing

Testing is run in GitHub actions, test actions can be listed with: `./go test`

## License

This project is licensed under the [BSD 3-Clause License](LICENSE.txt)
