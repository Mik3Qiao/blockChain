# Startup Guide:

* To bring the network online:
    ```bash
        ./startNetwork.sh
    ```

* To bring the network offline:
    ```bash
        ./shutdownNetwork.sh
    ```

If there are any problems bringing the network online try running shutdownNetwork then startNetwork again. That should fix any issues.

startNetwork is currently configured to go through an example of calculating a taskmatching to demonstrate how the chaincode is used to calculate taskmatchings. To remove this and perform your own manual tests simply go to the scripts folder and edit script.sh, scroll down to the bottom and you will see a line that says to comment out everything beyond that point to just start up the network without any chaincode being run. Examples of how to run the code manually
are found in the example_method_use file.

# Editing Chaincode:

The chaincode can be found in the chaincode folder, if it is modified the network won't actually update to the new chaincode unless you change the version number specified at the top of /scripts/utils.sh

# Additional Notes:

Currently a taskmatching should always be created with the name "work", this is because "work" is hardcoded into a few methods in the chaincode such as calculateTaskMatching(). This can be changed by adding additional parameters to the other methods or the ability to name a taskmatching can just be removed. 