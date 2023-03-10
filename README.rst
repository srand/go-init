

Yet another init
================

Example configuration:

  .. code:: yaml
    
    cgroups:
      - name: "services"
        config:
          cgroup.subtree_control: "+cpu +cpuset"
      
      - name: "services.throttled"
        config:
          cgroup.subtree_control: "+cpuset"
          cpu.max: 10000 # max 10 %
    
    modules:
      include:
        - /etc/modules-load.d/*.conf
      probe:
        - name: "md_mod"
          parameters: 
            - "start_ro=1"

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
    
      - name: stress
        command: ["stress-ng", "-c", "1"]
        cgroup:
          name: "services.throttled"
          config:
            # Pin service to cpu0
            cpuset.cpus: 0
    
    sysctl:
      include:
        - /etc/sysctl.d/*.conf
      parameters:
        - key: net.ipv4.conf.all.forwarding
          value: 1


Control Groups (cgroups)
------------------------

Process control groups are used to control system resources available to a service or task. 
For example, you may restrict processor time, memory consumption, io throughput, etc.

Schema
^^^^^^

+--------------+--------+--------------------------------------------------------------------------------------+
| Attribute    | Type   | Description                                                                          |
+==============+========+======================================================================================+
| name         | string | Name of the control group. '.' characters descend the control group hierachy.        |
+--------------+--------+--------------------------------------------------------------------------------------+
| config       | object | List of key-value pairs configuring the cgroup. Keys are passed directly to cgroups. |
+--------------+--------+--------------------------------------------------------------------------------------+
| config.key   | string | Attribute key, e.g. 'cpu.max'.                                                       |
+--------------+--------+--------------------------------------------------------------------------------------+
| config.value | string | Attribute value, e.g. 10000                                                          |
+--------------+--------+--------------------------------------------------------------------------------------+

Conditions
^^^^^^^^^^

 - cgroups.<cgroup-name>.configured


Modules
-------

Loads and monitors kernel modules. A condition is created for each loaded module, 
whether it is loaded by init or elsewhere. If a module is unloaded, the condition 
is deasserted. 


Schema
^^^^^^

+-------------------+----------+---------------------------------------------------+
| Attribute         | Type     | Description                                       |
+===================+==========+===================================================+
| include           | string   | Glob pattern matching module config files to load |
+-------------------+----------+---------------------------------------------------+
| probe             | []module | Modules to probe                                  |
+-------------------+----------+---------------------------------------------------+
| module.name       | string   | Name of module                                    |
+-------------------+----------+---------------------------------------------------+
| module.parameters | []string | List of module parameters                         |
+-------------------+----------+---------------------------------------------------+

Conditions
^^^^^^^^^^

 - modules.<module-name>.loaded


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
| cgroup          | object    |                                                   |
+-----------------+-----------+---------------------------------------------------+
| cgroup.name     | string    | Name of parent control group.                     |
+-----------------+-----------+---------------------------------------------------+
| cgroup.config   | object    | List of attributes applied to the service cgroup. |
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

+---------------+----------+------------------------------------------------+
| Attribute     | Type     | Description                                    |
+===============+==========+================================================+
| name          | string   | Name of the service                            |
+---------------+----------+------------------------------------------------+
| cgroup        | object   |                                                |
+---------------+----------+------------------------------------------------+
| cgroup.name   | string   | Name of parent control group.                  |
+---------------+----------+------------------------------------------------+
| cgroup.config | object   | List of attributes applied to the task cgroup. |
+---------------+----------+------------------------------------------------+
| command       | string   | Command to run and supervise                   |
+---------------+----------+------------------------------------------------+
| conditions    | []string | List of preconditions that must be met         |
+---------------+----------+------------------------------------------------+


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
