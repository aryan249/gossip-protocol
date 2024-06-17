# Message Tracker

## Major components of message tracker are as follows
    - P2P Node
    - Tracker
    - Processor

## How it works
> P2P node consist of a pubsub manager and a connection gater these are 2 major components for message propogation via p2p node

### Pubsub manager 
> This component is used to manager topics and there repective handlers to propogate the message in a correct way. It can also be used to cancel subscription to a topic and a subscript to a new topic with particular channels to handle message propogation for each topic.

### Connection gater 
> Id is used for managing peers connected and validate a meesage if it is send by a trusted validator. And to manage new connections.

### p2p message propogation works on 2 channels obsSendReq and ObsReceive Req
> When user wants to send a message he can propogate a message on obsSendReq and message will be propogated to all the peers.
> Similarly when some other peer publiuhes a message it is handled by obsReceiveReq.

## Tracker 
> It is the main component which works to transfer message recieved by node and creating it entry to data base and quering the data.
> Tracker struct has it's own database which keeps record of all the messages.
> Tracker implements all the methods which are required by Message Tracker interface.

## Processor 
> This go routine relays messages received by p2p node to message tracker
> When A message is receive by p2p node it send message to obsRecived channel
> Processor monitor's messages on this channel and Add a recived message to tracker which creates an entry in data base


## Config file components

### Logger level
> This defines the severity level of logs you want

### p2p-config
> This config pretty much defines the configuration for your p2p node these are the components
   >> ***listen-addresses*** - This is an array of addresses on which your node will listen to.  
   >> ***bootstrap_nodes*** - This is an array of bootstarp addresses of nodes which will be added to your peers.  
   >> ***discovery_methods*** - Used to define which tyope of discovery method one wants to use **MDNS** or **DHT**.  
   >>>**Note** - MDSN works on local machine but not on internel where as DHT does dht requires bootstrap addresses  so   
   later other nodes can discover the network with them.

### Node details
> This only consist of node oprators ***private_key*** which will be used as an identity for node

### Storage
> Storage defines configuration for postgres data base
   >> ***url*** - Database connection url.  
   >> ***host*** - Host on which db services are running.  
   >> ***port*** - Port for db services.  
   >> ***ssl_mode*** - To enable or disable ssl.  
   >> ***db_name*** - Name of database.  
   >> ***user*** - User who have access of database.  
   >> ***password*** - Postgres password.  

### Coverage on components is as follows.
> **MessageTracker** - 89 %  
> **P2P_Node** - 72 %  
> **Database** - 79 % 

## To test messaging one can shold follow these steps

> **Step 1** - Start the main node with ***go run .***  
> **Step 2** - Once the node is started go to cmd folder and run publish message with ***go run publish-message.go*** This will propogate new message.





