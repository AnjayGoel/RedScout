# Redmon - Redis Namespace Monitor

Redmon is a monitoring tool for redis that provides namespace-level insights of your Redis database. Unlike
traditional Redis monitoring tools that focus on server-level metrics, Redmon helps you understand which caches are
consuming the most memory, which namespaces are most frequently accessed etc.

## How It Works

Redmon combines three Redis commands to provide namespace-level insights:

- **`SCAN`** - Enumerates keys without blocking, groups by namespace
- **`MEMORY USAGE`** - Measures precise memory consumption per key
- **`MONITOR`** - Captures real-time operations and frequencies

## Requirements

- **Redis Version**: 4.0.0 or higher (required for `MEMORY USAGE` command)

## Installation

Download the appropriate binary for your platform from the releases page.

## Usage

```bash
./redmon \
  -host redis.example.com \
  -port 6379 \
  -password "your-password" \
  -tls true \
  -monitor-duration 30 \
  -scan-size 10000
```

## Configuration Options

### Connection Settings

| Flag        | Type   | Default   | Description                               |
|-------------|--------|-----------|-------------------------------------------|
| `-host`     | string | _(empty)_ | Redis host address                        |
| `-port`     | int    | `6380`    | Redis port number                         |
| `-password` | string | _(empty)_ | Redis password for authentication         |
| `-db`       | int    | `0`       | Redis database number to analyze          |
| `-tls`      | string | `false`   | Use TLS for Redis connection (true/false) |

### Analysis Settings

| Flag                | Type   | Default   | Description                                                   |
|---------------------|--------|-----------|---------------------------------------------------------------|
| `-delimiter`        | string | `:`       | Delimiter for separating Redis keys into namespaces           |
| `-scan-size`        | int    | `5000`    | Number of keys to scan                                        |
| `-monitor-duration` | int    | `10`      | Duration in seconds to run the `monitor` command              |
| `-refresh-interval` | int    | `5`       | Interval in seconds between Redis info refreshes              |
| `-id-regex`         | string | _(empty)_ | Space-separated list of regex patterns to infer IDs from keys |

### Output Settings

| Flag        | Type   | Default         | Description                                |
|-------------|--------|-----------------|--------------------------------------------|
| `-logs-dir` | string | _OS's temp dir_ | Directory to store temporary analysis logs |

## Notes

- Uses Redis [MONITOR](https://redis.io/docs/latest/commands/monitor/) command, so be careful when using in production
  environments
- Results are estimates based on sampling, not exhaustive key scanning