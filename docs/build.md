# Building `gatekeeper-securitycenter`

`gatekeeper-securitycenter` is provided as a binary command-line tool and a
container image.

## Building the binary

Build the command-line tool for your platform:

```sh
go install github.com/googlecloudplatform/gatekeeper-securitycenter@latest
```

## Building the container image

Build and publish a container image for the controller:

```sh
git clone https://github.com/GoogleCloudPlatform/gatekeeper-securitycenter.git

cd gatekeeper-securitycenter

export KO_DOCKER_REPO=gcr.io/$(gcloud config get-value core/project)

ko publish --base-import-paths --tags latest .
```

[`ko`](https://github.com/google/ko) is a command-line tool for building
container images from Go source code. It does not use a `Dockerfile` and it
does not require a local Docker daemon.

If you would like to use a different base image, edit the value of
`defaultBaseImage` in the file [`.ko.yaml`](../.ko.yaml).
