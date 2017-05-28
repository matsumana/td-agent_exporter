# td-agent_exporter

[![CircleCI](https://circleci.com/gh/matsumana/td-agent_exporter/tree/master.svg?style=shield)](https://circleci.com/gh/matsumana/td-agent_exporter/tree/master)

[td-agent](https://docs.treasuredata.com/articles/td-agent) exporter for [Prometheus](https://prometheus.io/)

# export metrics

- td_agent_cpu_time
- td_agent_resident_memory_usage
- td_agent_virtual_memory_usage

# Command Line Options

Name     | Description | Default | note
---------|-------------|----|----
-web.listen-address | Address on which to expose metrics and web interface | 9256 |
-web.telemetry-path | Path under which to expose metrics | /metrics |
-fluentd.process_name_prefix | fluentd's process_name prefix | | [Fluentd official documentation](http://docs.fluentd.org/v0.12/articles/config-file#processname)
-log.level | Log level | info |

# How to build

    $ make

# How to run unit test

    $ make unittest
