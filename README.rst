

Yet another init
================

Example configuration:

  .. code:: yaml

    tasks:
      - name: sshd.tmpdir
        command: ["mkdir", "-p", "/var/run/sshd"]

    services:
      - name: ntpd
        command: ["/usr/sbin/ntpd", "-n", "-N", "-p", "/var/run/ntpd.pid"]

      - name: rsyslogd
        command: ["/usr/sbin/rsyslogd", "-n"]

      - name: sshd
        command: ["/usr/sbin/sshd", "-D"]
        conditions:
          - tasks.sshd.tmpdir.state.completed

    sysctl:
      include:
        - /etc/sysctl.d/*.conf
      parameters:
        - key: net.ipv4.conf.all.forwarding
          value: 1


Services
--------

Services are supervised processes. They are automatically started once preconditions are met
and restarted if they fail.

Schema
^^^^^^

+-----------------+-----------+---------------------------------------------------+
| Attribute       | Type      | Description                                       |
+=================+===========+===================================================+
| name            | string    | Name of the service                               |
+-----------------+-----------+---------------------------------------------------+
| command         | string    | Command to run and supervise                      |
+-----------------+-----------+---------------------------------------------------+
| pidfile         | object    |                                                   |
+-----------------+-----------+---------------------------------------------------+
| pidfile.create  | bool      | Set true for services that don't create pidfiles  |
+-----------------+-----------+---------------------------------------------------+
| pidfile.path    | string    | Path to pidfile. Default. /var/run/<name>.pid     |
+-----------------+-----------+---------------------------------------------------+
| conditions      | []string  | List of preconditions that must be met            |
+-----------------+-----------+---------------------------------------------------+

Actions
^^^^^^^

 - services.<service-name>.action.start
 - services.<service-name>.action.stop


Conditions
^^^^^^^^^^

 - services.<service-name>.state.starting
 - services.<service-name>.state.stopping
 - services.<service-name>.state.running
 - services.<service-name>.state.stopped
 - services.<service-name>.state.error


System Control
--------------

Sets system control parameters.


Schema
^^^^^^

+-----------------+-------------+--------------------------------------------------+
| Attribute       | Type        | Description                                      |
+=================+=============+==================================================+
| include         | string      | Glob pattern matching property files to load     |
+-----------------+-------------+--------------------------------------------------+
| parameters      | []Parameter | List of parameters to assign                     |
+-----------------+-------------+--------------------------------------------------+
| Parameter.key   | string      | Configuration key                                |
+-----------------+-------------+--------------------------------------------------+
| Parameter.value | string      | Configuration value                              |
+-----------------+-------------+--------------------------------------------------+


Tasks
-----

Tasks are run once when preconditions have been met. They can be manually rerun upon manual request,
or by trigger.

Schema
^^^^^^

+-----------------+-----------+---------------------------------------------------+
| Attribute       | Type      | Description                                       |
+=================+===========+===================================================+
| name            | string    | Name of the service                               |
+-----------------+-----------+---------------------------------------------------+
| command         | string    | Command to run and supervise                      |
+-----------------+-----------+---------------------------------------------------+
| conditions      | []string  | List of preconditions that must be met            |
+-----------------+-----------+---------------------------------------------------+


Actions
^^^^^^^

 - tasks.<task-name>.action.start
 - tasks.<task-name>.action.stop


Conditions
^^^^^^^^^^

 - tasks.<task-name>.state.pending
 - tasks.<task-name>.state.running
 - tasks.<task-name>.state.stopping
 - tasks.<task-name>.state.completed
 - tasks.<task-name>.state.error
