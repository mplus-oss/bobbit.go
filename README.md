# bobbit
Simply "yet" UNIX Socket based job runner with more circus.

## Usage

Start the daemon:

```
bobbitd
```

Run a job:
```
bobbit create <job_name> -- <command>
```

Wait for a job:
```
bobbit wait <jobID|jobName>
```

Check job status:
```
bobbit status <jobID|jobName>
```

List jobs:
```
bobbit list
```

Stop a running job:
```
bobbit stop <jobID|jobName>
```

Tail job log:
```
# Stream log over socket
bobbit tail <jobID|jobName> -f

# Tail log file directly from filesystem (bypasses socket streaming)
bobbit tail <jobID|jobName> -f -d
```

## Configuration

Configurations are only possible with environment variables

- `BOBBIT_SOCKET_PATH` : Path to Socket, if directory doesn't exist, it will try to create it. (Default: `/tmp/bobbitd.sock`)
- `BOBBIT_DATA_DIR`: Data directory, it's important to know that both `bobbit` and `bobbitd` will use this directory. (Default: `/tmp/bobbitd`)

## Running inside OCI container

Use `tini`, or if you're using Docker, pass `--init` flag.
