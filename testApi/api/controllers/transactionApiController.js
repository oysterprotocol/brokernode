
//'use strict';

//refactor later
const IOTA = require('iota.lib.js');
const mysql = require('mysql2');
const Sequelize = require('sequelize');
const sequelize = new Sequelize('default', 'root', 'root', {
    host: 'localhost',
    dialect: 'mysql',
    define: {
        timestamps: true
    },
});

const iota = new IOTA({
    'provider': 'http://localhost:14265'
});

const Transactions = require('../../../peer-db/models/transactions.js')(sequelize, Sequelize);



confirm_work();

//create table of ids
exports.add_peer_id = function(req, res) {

  //get peer id
  var peer_id = req.query.peerid;

  //move into DB object
  var con = connect();

  //the way the table is set up we are required to manually add the timestamps.
  //laravel adds them automatically.  Currently not using.
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

//API CALL 2:  REQUEST TO START A TRANSACTION
exports.start_transaction = function(req, res) {
  var need = req.query.need;
  //FOR NOW WE JUST DO WEBNODES, SO WE GET THE LIST OF PEER IDS HERE.
  var date1 = new Date().toISOString().slice(0, 19).replace('T', ' ');
  var date2 = new Date().toISOString().slice(0, 19).replace('T', ' ');
  var con = connect();

  //add transaction and get txid
  var sql = "INSERT INTO default.Transactions (need_requested, createdAt, updatedAt) VALUES (\"" + need + "\",\"" +
  		date1 + "\",\""+ date2 + "\");";
  con.query( sql, function(err, result){
    //get txid


    tid = result.insertId;

    console.log("Created transaction with id ", tid);

    //get items
    var con2 = connect();

    //get list of peers to send  We are not hashing yet.
    //move this into the function outlined below
    //though we might move this all to go in which case we would want '
    //to go through it again real fast after defining the parameters and return values for different api 
    //calls
    
    var sql = "SELECT * FROM default.PeerIds;";
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


//API CALL 3:  TELL SELLER WHICH ITEM AND GET WORK
exports.item_selected = function(req, res) {

  //look up user in row
  var txid = req.query.txid;
  var ind = req.query.itemIndex;

  var con = connect();

  var sql = "SELECT * FROM default.Transactions WHERE id =\""+ txid + "\";";
  
  //var webnodes = getWebnodeAddresses();

  console.log(webnodes);
  console.log("here");

  con.query( sql, function(err, result){
    
    //we were dealing with the index of the need.  I want to change it so the web node passes the hash rather than
    //index though that also requires additional cpu cycles.
 
	//this is the need requested  LATER WE WILL SWITCH TO GET THE CUSTOMER'S LIST BASED ON ITEM TYPE.
    console.log(result[0].need_requested);

    //get webnode addresses
//    var con2 = connect();
//
//    var sql = "SELECT * FROM default.PeerIds;";
//    var webnode_array = [];
//    con2.query( sql, function(err, result){
//      //get txid
//      console.log("listing webnodes " );
//
//      console.log(result);
//      result.forEach(function(element) {
//         webnode_array.push(element.peer_id);
//      });
//
//      //console.log(result);
//      console.log(webnode_array);
    //  console.log(result.insertId.toString());

      
    //we then update the SQL
    
    
    var update_transaction_sql = "UPDATE default.Transactions SET item_selected_index = \""+ ind + "\" WHERE transaction_id = "+txid+";"
    
    var another_connection = connect();
    
    another.connection.query(update_transaction_sql, function(err, result){
    	
    	//another callback!
    	console.log("Purchaser has selected an item.  The transaction has been updated.");
    	
    	//TODO: GET AND RETURN SOME WORK 
    	res.send({ message: 'somemessage', address: 'testaddress'});
    	});
      
    });


//var confirm_work = function(req, res) {
function confirm_work() {

    // //look up user in row
    // let txFromWebNode = {
    //     // transaction_id: req.query.txid,
    //     // address: req.query.address,
    //     // message: req.query.message
    //
    //     transaction_id: 'cd6e7190-06d1-4c70-9593-32cda3134ab3',
    //     address: 'ABSCUCQCXAPCCBUAZAXATCCBSCZAZATCCBTCSCABTCPCYAUCRCRCTCABZATCYAYABBZABBSCRCQCXAQCR',
    //     message: 'DXCADSCACVAPBWCPAVABDWAUBRCEDACNBZANBWCABCDZCZCTCWBCCKBFDBCBDRBDCKDEDVCPBTBICDCQCWCXBLDEDNDTCYBQBNDJDUCNBUBQCPAHCACRC9DTAICWATBRCVBEDRBWCFCVCPCUBTBEDSCLBYAGCCDTANBZBBCCCCCKDICLBZBADICSCTBBDYBYBDCDDACTCTBRBCCHCUCCDJDTBECMDVCXBNDADCCECWCHCSCWBUCZBUAVBRBTAACBBNB9CVCSBWBED9DIDLBNBYCUBPAZBKBUCMDNDZAGDNBPCYCXBVCHDGCUAGCPAUAABMBBBQCDDSBWBBDUAFCQB9C9BACZCGCWBKBFCRBBBACZCQCPAXBTACDADADYC9CBBGCRCUADCCBZCPAADNBFC9CKBZCLDVAICHCCBAC9BXCRCXCGDICNDYCNBHCZCOBEDZCUCWAJDQCUBPAYCPBADDDRBKBYCXAPC9DUAPADDZATAZBCCMBPAICZAZCPCWCZAUBKBMDUCGCMBHCWCYBXBPCNDTCTBEDKBHCWAZAMDJDYBYCQCTAJDYCWARCABFDQCLDYAZBUCSC9DMD9CXATBBDWANBACZBNBWATCXAPB9DYCPBJDADXBADGCTBECUCNDTAYBUCFCGDVAADBCFCLDABSBRC9BPCYCBDLBBDDCYBZBIDPAYBOBIDEDXADDICPBQBICWCTAWCNDNBRB9BPAJDLBTBECADCBRCECFCABRBUCKBICZARCOBZANBECBBCBWBZBHCUAMDWAPCFCWCJDDCWCLBICDCUBPASBHDABHCPCXAZAFCEDDCCBPBKBZBCCOBVBIDCCXCBCFCMDSBLBIDZCOBMDSCBBMDYAKDQBXBZCSBXBSCVAQBVCVAYAPBACECNDQCLBZBJDCDPCMBWCIDLBZCNDDDWBLDDDKDFDXBACPCHDYBABSCXBADPANBVBUBWCHCFDLDOBACBDYCTBHDWAVBNBLBBDKBUCGCQCYCPCTBZCKBDCCBZCMBECADQBZCECTATAUCLBPCWAZAXBADPAHDDCACCDXAECGCKBRBCBKBXBYATCSBECFDMDZBLD9BGCFCCBFCVAICYCHCWACCNDSBYADDCCCDECEDNDECTAVCMDYARCACUATBFDFDABCBQCSCGCWCXCBCPAMBWCICTCFDKDSCRBNBVCACUBBDBCZBPBADJDZACCHCIDYCUABBHDJDMDRCECBDXAXBSBCBWAYBKBFCDCUBZCVBHCBCDDWCYAKBRBKDUAEDTBACABDDZAVBYBSCECDCEDBBICRCECRCXBACRBECUATBBBICLDGD9CRCUBYABBOBXBVCPBKDPCWCBBTCKBLBKDHDHCQBBCGCNBSBXBGCYCABKDJDXCVCJDCDVBYBOBSCLBPAUC9BYBAC9BZBABVADDPCPBSBNB9DUCZAFCYAZBUBHCUATAFDGCYBWBCCFC9DKDNBJDCCVBRCLDCDICPCBBNBHDKBADWADDOBKBMBRBZCFCWBLDNDABOBYBSCTCYCYBCDBDRBUCVARCCDCDVCFDQCHDACFDFCYCMDEDZCOBDC9BPCQCBDOBVBMBXBPCVCVCWCYBXCEDLBICTCDDUBLBWAGDWCABPCECLDZA9CVCTAGCXBFDRCPCQCVAFDCCXA9BVCWCLBCDADADADHDYAMBGCPBCBLBWACBHDCDIDRCYBGDTBCBZCWAXAADWCTC9BKDECNDCBLDPCECGCPCBBBDYCVCQBIDXCCDKB9BABTBVCCBWCFCCBBCWBVAPAEDXBJDGDWAVCDDSCWBYBZADCEDNDZCYCIDFCHCKBOBHCXCKDOBCBWCKBOBRBZCYBFDPBVCUACBCCOBLBTBBC9BJDMDBCTAPATBADZAGCMBPBSBXCTBJDHDNDGCBCLBOB9BND9DADXATBLBRCXBXBDDSBWBBCCDYBJDCBNDYBSBTCZBPBGC9CMDMBRCCD9BXC9BACPCPADDEDXAHCTCNBWCUAHCRCGDVBWCTBBDZBXAPBWABDRCSCFD9DWAUAXBUALBSCTBYC9DUALDZCCBQBJDUCXCZAQBTBRBJDBDGCDDICGDBCNBCDLBMBFDKDACMDDCKDMDKDDDUAVCVCYBDDCBACWCBDHCBDBBKDXCIDZBADWBOBQBQCND9BECYAPAEDVA9BGCLDICNDCDVASBPC'
    // };
    //
    // let txInDB = {
    //     transaction_id: '',
    //     address: '',
    //     message: ''
    // };
    //
    // return Transactions.findOne({
    //     where: {
    //         transaction_id: txFromWebNode.transaction_id
    //     }
    // }).then(result => {
    //     txInDB.transaction_id = result.get('transaction_id');
    //     txInDB.address = result.get('address');
    //     txInDB.message = result.get('message');
    //     return check_tangle_for_work(txInDB);
    // });
};

function check_tangle_for_work(txInDB) {

    let searchValues = {
        addresses: [txInDB.address]
    };

    iota.api.findTransactionObjects(searchValues, function(error, result) {
        console.log(result);
    })
}

exports.confirm_work = confirm_work;


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

//use this in the functions above
function getWebnodeAddresses(){
  var con2 = connect();

  //add transaction and get txid
  var sql = "SELECT * FROM default.PeerIds;";
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
