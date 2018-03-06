
//'use strict';

//refactor later
var mysql = require('mysql');
var mariaSQL = require('mariasql');
const Sequelize = require('sequelize');
const sequelize = new Sequelize('default', 'root', 'root', {
    host: 'localhost',
    dialect: 'mysql',
    define: {
        timestamps: true
    },
});

//create table of ids
exports.add_peer_id = function(req, res) {

  //get peer id
  var peer_id = req.query.peerid;

  var con = connect();

  
  var date1 = new Date().toISOString().slice(0, 19).replace('T', ' ');
  var date2 = new Date().toISOString().slice(0, 19).replace('T', ' ');
  //add peer sql
  var sql = "INSERT INTO default.PeerIds (peer_id, createdAt, updatedAt) VALUES (\"" + peer_id + "\",\"" +
  		date1 + "\",\""+ date2 + "\");";
  con.query( sql, function(err, result){
	  console.log(err);
    console.log("Added new peer id.");
    res.send("accepted");
  });
};

var tid = -1;


exports.start_transaction = function(req, res) {
  var need = req.query.need;
  //FOR NOW WE JUST DO WEBNODES, SO WE GET THE LIST OF PEER IDS HERE.

  var con = connect();

  //add transaction and get txid
  var sql = "INSERT INTO testdb.transactions (need_requested) VALUES (\"" + need + "\");";
  con.query( sql, function(err, result){
    //get txid


    tid = result.insertId;

    console.log("Created transaction with id ", tid);

    //get items
    var con2 = connect();

    //add transaction and get txid
    var sql = "SELECT * FROM testdb.peer_ids;";
    var webnode_array = [];
    con2.query( sql, function(err, result){
      //get txid
      console.log("listing webnodes " );

      console.log(result);
      result.forEach(function(element) {
         webnode_array.push(element.peer_id);
      });

      //console.log(result);
      console.log(webnode_array);
    //  console.log(result.insertId.toString());

      res.send({ txid: tid, items: webnode_array});
      //return webnode_array;
    });
    console.log("possibleWebnodes");

    //console.log(possibleWebnodes);

  });


};


exports.need_selected = function(req, res) {

  //look up user in row
  var txid = req.query.txid;
  var ind = req.query.itemIndex;

  var con = connect();

  var sql = "SELECT * FROM testdb.transactions WHERE id =\""+ txid + "\";";

  var webnodes = getWebnodeAddresses();

  console.log(webnodes);
  console.log("here");

  con.query( sql, function(err, result){
    //get txid
    console.log("Need has been selected");

    console.log(sql);

    console.log(result[0].need_requested);

    //get webnode ADDRESSES
    var con2 = connect();

    //add transaction and get txid
    var sql = "SELECT * FROM testdb.peer_ids;";
    var webnode_array = [];
    con2.query( sql, function(err, result){
      //get txid
      console.log("listing webnodes " );

      console.log(result);
      result.forEach(function(element) {
         webnode_array.push(element.peer_id);
      });

      //console.log(result);
      console.log(webnode_array);
    //  console.log(result.insertId.toString());

      res.send(webnode_array[ind]);
      //return webnode_array;
    });


  });

//FOR NOW WE ARE ONLY DOING WEBNODE ADDRESSES SO WE DON'T NEED TO ADD.
//I AM GOING TO SEND THE INDEX OF THE ITEM FOR NOW
};


exports.confirm_work = function(req, res) {

    //look up user in row
    var txid = req.query.txid;

    console.log("in confirm_work");
};


function connect(){
  var con = mysql.createConnection({
      host: "127.0.0.1",
      port:  3306,
      user: "root",
      password: "root",
      db: "default"

  });
  return con;
}

function getWebnodeAddresses(){
  var con2 = connect();

  //add transaction and get txid
  var sql = "SELECT * FROM testdb.peer_ids;";
  var webnode_array = [];
  con2.query( sql, function(err, result){
    //get txid
    console.log("listing webnodes " );

    console.log(result);
    result.forEach(function(element) {
       webnode_array.push(element.peer_id);
    });

    //console.log(result);
    console.log(webnode_array);
  //  console.log(result.insertId.toString());

    return(webnode_array);
    //return webnode_array;
  });
}
