

Yet another init
================

Example configuration:

  .. code:: yaml
    
    services:
    - name: "sleep.1"
      command: ["sleep", "100"]
      pidfile: 
      create: true
      conditions:
        - services.fail.state.running
        - services.sleep.2.state.running
    - name: "sleep.2"
      command: ["sleep", "10"]
      pidfile: 
        create: true
    - name: "fail"
      command: ["sh", "-c", "sleep 5 && false"]
      pidfile: 
        create: true
      triggers:
        - name: "start when stopped"
          conditions:
            - services.fail.state.stopped
          actions:
            - services.fail.action.start


Actions
-------

 - services.<service-name>.action.start
 - services.<service-name>.action.stop


Conditions
----------

 - services.<service-name>.state.starting
 - services.<service-name>.state.stopping
 - services.<service-name>.state.running
 - services.<service-name>.state.stopped
 - services.<service-name>.state.error
