//refactor later
const IOTA = require('iota.lib.js');
const mysql = require('mysql2');
const uuidv4 = require('uuid/v4');
const Sequelize = require('sequelize');
const sequelize = new Sequelize('default', 'root', 'root', {
    host: 'localhost',
    dialect: 'mysql',
    define: {
        timestamps: true
    },
});
const Raven = require('raven');
const Analytics = require('analytics-node');
const analytics = new Analytics(process.env.SEGMENT_WRITE_KEY, {flushAt: 1});
//flushAt means sending a request every 1 message
//we can change this

const iota = new IOTA({
    'provider': 'http://localhost:14265'
});

Raven.config(
    process.env.SENTRY_DSN
).install();

const Transactions = require('../../../peer-db/models/transactions.js')(sequelize, Sequelize);

//API CALL 1:  ADD A PEER ID.
exports.add_peer_id = function (req, res) {

    //get peer id
    let peer_id = req.body.peerid;

    //move into DB object
    let con = connect();

    //the way the table is set up we are required to manually add the timestamps.
    //laravel adds them automatically.  Currently not using.
    let date1 = new Date().toISOString().slice(0, 19).replace('T', ' ');
    let date2 = new Date().toISOString().slice(0, 19).replace('T', ' ');

    let id = uuidv4();

    //add peer sql
    let sql = "INSERT INTO default.PeerIds (id, peer_id, createdAt, updatedAt) VALUES (\"" + id + "\",\"" + peer_id + "\",\"" +
        date1 + "\",\"" + date2 + "\");";
    con.query(sql, function (err, result) {

        if (err) {
            Raven.captureException(err);
        }

        console.log(err);
        console.log("Added new peer id.");

        analytics.track({
            userId: req.headers.host.split(":")[0],
            event: 'add_peer_id',
            properties: {
                request_origin: req.headers.origin,
                peer_id: peer_id,
            }
        });

        res.send("accepted");
    });
};

let tid = -1;

//API CALL 2:  REQUEST TO START A TRANSACTION
exports.start_transaction = function (req, res) {
    let need = req.body.need_requested;

    //FOR NOW WE JUST DO WEBNODES, SO WE GET THE LIST OF PEER IDS HERE.
    let date1 = new Date().toISOString().slice(0, 19).replace('T', ' ');
    let date2 = new Date().toISOString().slice(0, 19).replace('T', ' ');
    let con = connect();

    let id = uuidv4();
    //add transaction and get txid
    let sql = "INSERT INTO default.Transactions (transaction_id, need_requested, createdAt, updatedAt, item_selected_index) VALUES (\"" + id + "\",\"" + need + "\",\"" +
        date1 + "\",\"" + date2 + "\",\"-1\");";

    con.query(sql, function (err, result) {
        //get txid

        if (err) {
            Raven.captureException(err);
        }

        console.log(err);
        console.log(result);
        tid = id;

        console.log("Created transaction with id ", tid);

        //get items
        let con2 = connect();

        //get list of peers to send  We are not hashing yet.
        //move this into the function outlined below
        //though we might move this all to go in which case we would want '
        //to go through it again real fast after defining the parameters and return values for different api
        //calls

        let sql = "SELECT * FROM default.PeerIds;";
        let webnode_array = [];
        con2.query(sql, function (err, result) {
            //get txid

            if (err) {
                Raven.captureException(err);
            }

            console.log("listing webnodes ");

            console.log(result);
            result.forEach(function (element) {
                webnode_array.push(element.peer_id);
            });

            //console.log(result);
            console.log(webnode_array);
            //  console.log(result.insertId.toString());

            analytics.track({
                userId: req.headers.host.split(":")[0],
                event: 'start_transaction',
                properties: {
                    request_origin: req.headers.origin,
                    need: need,
                    tx_id: tid,
                    webnode_array: webnode_array
                }
            });

            res.send({txid: tid, items: webnode_array});
        });
        console.log("possibleWebnodes");
    });
};


//API CALL 3:  TELL SELLER WHICH ITEM AND GET WORK
exports.item_selected = function (req, res) {

    //look up user in row
    let txid = req.body.txid;
    let ind = req.body.itemIndex;

    let con = connect();

    let sql = "SELECT * FROM default.Transactions WHERE transaction_id =\"" + txid + "\";";

    //let webnodes = getWebnodeAddresses();

    con.query(sql, function (err, result) {

        if (err) {
            Raven.captureException(err);
        }

        let update_transaction_sql = "UPDATE default.Transactions SET item_selected_index = \"" + ind + "\" WHERE transaction_id = \"" + txid + "\";";

        let another_connection = connect();

        another_connection.query(update_transaction_sql, function (err, result) {

            if (err) {
                Raven.captureException(err);
            }

            console.log("Purchaser has selected an item.  The transaction has been updated.");

            iota.api.getTransactionsToApprove(4, undefined, function (err, result) {
                if (err === undefined) {

                    result.address = 'SEWOZSDXOVIURQRBTBDLQXWIXOLEUXHYBGAVASVPZ9HBTYJJEWBR9PDTGMXZGKPTGSUDW9QLFPJHTIEQ';
                    result.message = 'THISCANSAYANYTHING';
                    result.broadcastingNodes = [
                        '13.124.107.48',
                        '35.183.23.179',
                        '35.178.32.118',
                        '54.168.83.160',
                        '54.95.4.132'
                    ];
                    result.request_origin = req.headers.origin;
                    result.tx_id = txid;
                    result.item_index = ind;

                    analytics.track({
                        userId: req.headers.host.split(":")[0],
                        event: 'item_selected',
                        properties: result
                    });

                    //TODO: GET SOME WORK FROM THE DATA MAP.
                    res.send(result);
                } else {
                    Raven.captureException(err);
                }
            });
        });

    });

};

exports.report_work_finished = function (req, res) {

    //TODO Confirm work is done.

    let txid = req.body.txid;
//	  
    let con = connect();
//
    let sql = "SELECT * FROM default.Transactions WHERE transaction_id =\"" + txid + "\";";
//	  
//	
    con.query(sql, function (err, result) {

        if (err) {
            Raven.captureException(err);
        }

        //we were dealing with the index of the need.  I want to change it so the web node passes the hash rather than
        //index though that also requires additional cpu cycles.

        //this is the need requested  LATER WE WILL SWITCH TO GET THE CUSTOMER'S LIST BASED ON ITEM TYPE.
        let need_type = result[0].need_requested;
        let item_selected_index = result[0].item_selected_index;
        let webnode_array = [];
//		    items = null;
//		    
//		    //TODO:  Add other item types, for now we can sell other webnode addresses
//		    //this means that each time someone logs in everyone else can purchase their items.
        //switch(need_type){
        //case "webnode_address":
        let connection = connect();

        let sql = "SELECT * FROM default.PeerIds;";

        connection.query(sql, function (err, result) {

            if (err) {
                Raven.captureException(err);
            }

            result.forEach(function (element) {
                webnode_array.push(element.peer_id);
            });

            //return(webnode_array);
            let item = webnode_array[item_selected_index];

            analytics.track({
                userId: req.headers.host.split(":")[0],
                event: 'report_work_finished',
                properties: {
                    request_origin: req.headers.origin,
                    tx_id: txid,
                    item_index: item_selected_index,
                    need: need_type,
                    item: item
                }
            });

            res.send(item);

        });
        //}
//		      

//		    
//		    
//		    let update_transaction_sql = "UPDATE default.Transactions SET transaction_status  = \"TRANSACTION_COMPLETE\" WHERE transaction_id = "+txid+";"
//		    
//		    let another_connection = connect();
//		    
//		    //clunky programming,refactor into some sort of await thing.
//		    another_connection.query(update_transaction_sql, function(err, result){
//		    	
//		    	console.log("Purchaser has finished work.  The item is being sent.");
//		    	
//		    	//Send item
//		    	res.send(item);
//		    	});
//		      
//		    });


    });
};


function connect() {
    let con = mysql.createConnection({
        host: "127.0.0.1",
        port: 3306,
        user: "root",
        password: "root",
        db: "default"

    });
    return con;
}

function getWebnodeAddresses() {
    let connection = connect();

    let sql = "SELECT * FROM default.PeerIds;";

    let webnode_array = [];

    connection.query(sql, function (err, result) {

        result.forEach(function (element) {
            webnode_array.push(element.peer_id);
        });

        if (err) {
            Raven.captureException(err);
        }

        analytics.track({
            userId: req.headers.host.split(":")[0],
            event: 'getWebnodeAddresses',
            properties: {
                request_origin: req.headers.origin,
                webnode_array: webnode_array
            }
        });

        return (webnode_array);

    });
}

//function getWorkFromDatamap(){
//	
//	return { address: "SEWOZSDXOVIURQRBTBDLQXWIXOLEUXHYBGAVASVPZ9HBTYJJEWBR9PDTGMXZGKPTGSUDW9QLFPJHTIEQ", message: "THISCANSAYANYTHING" }
//}
