
<a name="v0.9.1"></a>
## [v0.9.1](https://github.com/pgulb/plasma/compare/v0.9.0...v0.9.1) (2025-08-18)

### Fix

* **core:** add qemu and buildx actions to allow building multiarch docker image


<a name="v0.9.0"></a>
## [v0.9.0](https://github.com/pgulb/plasma/compare/v0.8.1...v0.9.0) (2025-08-18)

### Feat

* **core:** build and publish cli binaries also for arm64


<a name="v0.8.1"></a>
## [v0.8.1](https://github.com/pgulb/plasma/compare/v0.8.0...v0.8.1) (2025-08-18)

### Fix

* **cli:** fix panic on checking server version


<a name="v0.8.0"></a>
## [v0.8.0](https://github.com/pgulb/plasma/compare/v0.7.0...v0.8.0) (2025-08-18)

### Feat

* **cli:** check if plasma cli and server versions match before requests
* **cli:** when no container name provided for 'logs' command, get logs for plasma-server
* **cli:** add common 'help' commands to print usage
* **cli:** when using 'serve' or 'destroy', create compose file in temp dir instead of '.'
* **db:** add '<project_name>_' prefix to containers to allow same base names for different projects
* **server:** return server version.Version at /version

### Fix

* **cli:** show unknown command error message when using non defined command
* **db:** add prefix '<project_name>_' to mounts of type 'volume', otherwise when containers were created, another volume without project prefix was created


<a name="v0.7.0"></a>
## [v0.7.0](https://github.com/pgulb/plasma/compare/v0.6.2...v0.7.0) (2025-08-18)

### Feat

* **core:** more complete Task for tag creation
* **core:** rename package due to moving to personal repo

### Fix

* **grpc:** add '........<grpc-error> ' to error message returned from grpc logs service to avoid it being cut on the beggining


<a name="v0.6.2"></a>
## [v0.6.2](https://github.com/pgulb/plasma/compare/v0.6.1...v0.6.2) (2025-08-17)

### Fix

* **core:** change build server workflow to run on tag


<a name="v0.6.1"></a>
## [v0.6.1](https://github.com/pgulb/plasma/compare/v0.6.0...v0.6.1) (2025-08-17)

### Fix

* **core:** allow docker build job to push images


<a name="v0.6.0"></a>
## [v0.6.0](https://github.com/pgulb/plasma/compare/v0.5.0...v0.6.0) (2025-08-17)

### Feat

* **cli:** on 'plasma serve' fetch corresponding image from github
* **cli:** if 'develop' version.Version, build plasma-server locally on 'plasma serve'
* **core:** add building and pushing plasma-server docker image to github


<a name="v0.5.0"></a>
## [v0.5.0](https://github.com/pgulb/plasma/compare/v0.4.0...v0.5.0) (2025-08-17)

### Feat

* **core:** Add workflow to build and publish plasma cli


<a name="v0.4.0"></a>
## [v0.4.0](https://github.com/pgulb/plasma/compare/v0.3.0...v0.4.0) (2025-08-17)

### Feat

* **cli:** add 'logs' command to stream logs from gRPC
* **core:** run gRPC server to stream container logs
* **grpc:** run streaming logs service


<a name="v0.3.0"></a>
## [v0.3.0](https://github.com/pgulb/plasma/compare/v0.2.0...v0.3.0) (2025-08-16)

### Chore

* **core:** tidy go.mod
* **readme:** add link to changelog

### Feat

* **cli:** print colored Usage, add info in Red when missing parameters
* **cli:** more concise ps output, add data about mount count and ports mapped
* **container:** add support for environment vars

### Fix

* **db:** do not insert empty {} environment
* **server:** do not check if ctr.State == nil if ctr == nil


<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/pgulb/plasma/compare/v0.1.1...v0.2.0) (2025-08-16)

### Feat

* **container:** add support for port mapping
* **container:** add mounting volumes and bind mounts into containers (for now only rw mode)
* **db:** serialize port configuration into sqlite
* **db:** serialize volume configuration into sqlite

### Fix

* **cli:** mount docker.sock into plasma deployed by 'plasma serve'


<a name="v0.1.1"></a>
## [v0.1.1](https://github.com/pgulb/plasma/compare/v0.1.0...v0.1.1) (2025-08-15)

### Chore

* **task:** rename tasks to leverage namespaces, sort tasks
* **task:** hide internal tasks
* **task:** remove unused tasks

### Fix

* **server:** add nil check for ctr when getting status for /ps
* **server:** check for nil State of container when listing for /ps


<a name="v0.1.0"></a>
## v0.1.0 (2025-08-15)

### Feat

* **core:** Initial version

