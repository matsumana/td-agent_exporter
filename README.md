# td-agent_exporter

[![CircleCI](https://circleci.com/gh/matsumana/td-agent_exporter/tree/master.svg?style=shield)](https://circleci.com/gh/matsumana/td-agent_exporter/tree/master)

[td-agent](https://docs.treasuredata.com/articles/td-agent) exporter for [Prometheus](https://prometheus.io/)

# exported metrics

- td_agent_cpu_time
- td_agent_resident_memory_usage
- td_agent_virtual_memory_usage
- td_agent_up

# command line options

Name     | Description | Default | note
---------|-------------|----|----
web.listen-address | Address on which to expose metrics and web interface | 9256 |
web.telemetry-path | Path under which to expose metrics | /metrics |
fluentd.process_file_name | fluentd's process file name. | ruby | For example, td-agent is being executed from /opt/td-agent/embedded/bin/fluentd, specify "fluentd".
fluentd.process_name_prefix | fluentd's process_name prefix | | see also: [Fluentd official documentation](http://docs.fluentd.org/v0.12/articles/config-file#processname)
log.level | Log level | info |

# How to configure

td-agent_exporter find td-agent processes to collect metrics.  
If you use `process_name` in `td-agent.conf` like the following, please use `fluentd.process_name_prefix` option for td-agent_exporter.

## If td-agent has `process_name` setting

example setting of td-agent and its process name.

- td-agent.conf
  ```
  <system>
    process_name foo_1
  </system>

  <match debug.**>
    @type stdout
  </match>

  <source>
    @type forward
    port 24224
  </source>
  ```

- td-agent processes

  ```
  UID        PID  PPID  C STIME TTY          TIME CMD
  root      2489     1  0 07:07 ?        00:00:00 supervisor:foo_1
  root      2492  2489  0 07:07 ?        00:00:00 worker:foo_1
  ```

- Option for td-agent_exporter

  In this case, `fluentd.process_name_prefix` is required for td-agent_exporter like the following.

  ```
  /path/to/td-agent_exporter -fluentd.process_name_prefix=foo
  ```

- Exported metrics example

  __td-agent's process name is be used as `id` label.__

  ```
  # HELP td_agent_cpu_time td-agent cpu time
  # TYPE td_agent_cpu_time counter
  td_agent_cpu_time{id="foo_1"} 0.06
  # HELP td_agent_resident_memory_usage td-agent resident memory usage
  # TYPE td_agent_resident_memory_usage gauge
  td_agent_resident_memory_usage{id="foo_1"} 2.8913664e+07
  # HELP td_agent_virtual_memory_usage td-agent virtual memory usage
  # TYPE td_agent_virtual_memory_usage gauge
  td_agent_virtual_memory_usage{id="foo_1"} 1.9724288e+08
  # HELP td_agent_up the td-agent processes
  # TYPE td_agent_up gauge
  td_agent_up 1
  ```

## If td-agent doesn't have `process_name` setting

__In this case, don't need to use `fluentd.process_name_prefix`.__

example setting of td-agent __without__ process_name

- td-agent.conf

  ```
  <match debug.**>
    @type stdout
  </match>

  <source>
    @type forward
    port 24224
  </source>
  ```

- td-agent processes

  ```
  UID        PID  PPID  C STIME TTY          TIME CMD
  root      2450     1  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid --config /etc/td-agent/td-agent.conf
  root      2453  2450  0 07:07 ?        00:00:00 /opt/td-agent/bin/ruby /opt/td-agent/bin/fluentd --log /var/log/td-agent/td-agent.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent.pid --config /etc/td-agent/td-agent.conf
  ```

- Exported metrics example

  ```
  # HELP td_agent_cpu_time td-agent cpu time
  # TYPE td_agent_cpu_time counter
  td_agent_cpu_time{id="default"} 0.06
  # HELP td_agent_resident_memory_usage td-agent resident memory usage
  # TYPE td_agent_resident_memory_usage gauge
  td_agent_resident_memory_usage{id="default"} 2.895872e+07
  # HELP td_agent_virtual_memory_usage td-agent virtual memory usage
  # TYPE td_agent_virtual_memory_usage gauge
  td_agent_virtual_memory_usage{id="default"} 1.97251072e+08
  # HELP td_agent_up the td-agent processes
  # TYPE td_agent_up gauge
  td_agent_up 1
  ```

# How to build

```
$ make
```

# How to run unit test

```
$ make unittest
```
