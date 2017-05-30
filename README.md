# td-agent_exporter

[![CircleCI](https://circleci.com/gh/matsumana/td-agent_exporter/tree/master.svg?style=shield)](https://circleci.com/gh/matsumana/td-agent_exporter/tree/master)

[td-agent](https://docs.treasuredata.com/articles/td-agent) exporter for [Prometheus](https://prometheus.io/)

# export metrics

- td_agent_cpu_time
- td_agent_resident_memory_usage
- td_agent_virtual_memory_usage
- td_agent_up

# command line options

Name     | Description | Default | note
---------|-------------|----|----
web.listen-address | Address on which to expose metrics and web interface | 9256 |
web.telemetry-path | Path under which to expose metrics | /metrics |
fluentd.process_name_prefix | fluentd's process_name prefix | | see also: [Fluentd official documentation](http://docs.fluentd.org/v0.12/articles/config-file#processname)
log.level | Log level | info |

# Usage

## In case of use `process_name` within a `td-agent.conf`

e.g.

- td-agent_1.conf

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

- td-agent_2.conf

  ```
  <system>
    process_name foo_2
  </system>

  <match debug.**>
    @type stdout
    port 24225
  </match>

  <source>
    @type forward
  </source>
  ```

- processes

  ```
  UID        PID  PPID  C STIME TTY          TIME CMD
  root      2489     1  0 07:07 ?        00:00:00 supervisor:foo_1
  root      2492  2489  0 07:07 ?        00:00:00 worker:foo_1
  root      2502     1  0 07:07 ?        00:00:00 supervisor:foo_2
  root      2505  2502  0 07:07 ?        00:00:00 worker:foo_2
  ```

- command line option

  In this case, `fluentd.process_name_prefix` is required.

  ```
  /path/to/td-agent_exporter -fluentd.process_name_prefix=foo
  ```

The following metrics are exported:

  ```
  # HELP td_agent_cpu_time td-agent cpu time
  # TYPE td_agent_cpu_time counter
  td_agent_cpu_time{id="foo_1"} 0.06
  td_agent_cpu_time{id="foo_2"} 0.06
  # HELP td_agent_resident_memory_usage td-agent resident memory usage
  # TYPE td_agent_resident_memory_usage gauge
  td_agent_resident_memory_usage{id="foo_1"} 2.8913664e+07
  td_agent_resident_memory_usage{id="foo_2"} 2.891776e+07
  # HELP td_agent_up the td-agent processes
  # TYPE td_agent_up gauge
  td_agent_up 2
  # HELP td_agent_virtual_memory_usage td-agent virtual memory usage
  # TYPE td_agent_virtual_memory_usage gauge
  td_agent_virtual_memory_usage{id="foo_1"} 1.9724288e+08
  td_agent_virtual_memory_usage{id="foo_2"} 1.97156864e+08
  ```

## In case of don't use `process_name` within a `td-agent.conf`

e.g.

- td-agent_1.conf

  ```
  <match debug.**>
    @type stdout
  </match>

  <source>
    @type forward
    port 24224
  </source>
  ```

- td-agent_2.conf

  ```
  <match debug.**>
    @type stdout
    port 24225
  </match>

  <source>
    @type forward
  </source>
  ```

- processes

  ```
  UID        PID  PPID  C STIME TTY          TIME CMD
  root      2450     1  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf
  root      2453  2450  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_1.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_1.pid --config /etc/td-agent/td-agent_1.conf
  root      2463     1  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf
  root      2466  2463  0 07:07 ?        00:00:00 /opt/td-agent/embedded/bin/ruby /usr/sbin/td-agent --log /var/log/td-agent/td-agent_2.log --use-v1-config --group td-agent --daemon /var/run/td-agent/td-agent_2.pid --config /etc/td-agent/td-agent_2.conf
  ```

The following metrics are exported:

  ```
  # HELP td_agent_cpu_time td-agent cpu time
  # TYPE td_agent_cpu_time counter
  td_agent_cpu_time{id="td-agent_1"} 0.06
  td_agent_cpu_time{id="td-agent_2"} 0.06
  # HELP td_agent_resident_memory_usage td-agent resident memory usage
  # TYPE td_agent_resident_memory_usage gauge
  td_agent_resident_memory_usage{id="td-agent_1"} 2.895872e+07
  td_agent_resident_memory_usage{id="td-agent_2"} 2.891776e+07
  # HELP td_agent_up the td-agent processes
  # TYPE td_agent_up gauge
  td_agent_up 2
  # HELP td_agent_virtual_memory_usage td-agent virtual memory usage
  # TYPE td_agent_virtual_memory_usage gauge
  td_agent_virtual_memory_usage{id="td-agent_1"} 1.97251072e+08
  td_agent_virtual_memory_usage{id="td-agent_2"} 1.97185536e+08
  ```

# How to build

```
$ make
```

# How to run unit test

```
$ make unittest
```
