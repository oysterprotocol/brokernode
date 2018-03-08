
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


//var confirm_work = function(req, res) {
function confirm_work() {

    //look up user in row
    let txFromWebNode = {
        // transaction_id: req.query.txid,
        // address: req.query.address,
        // message: req.query.message

        transaction_id: 'cd6e7190-06d1-4c70-9593-32cda3134ab3',
        address: 'ABSCUCQCXAPCCBUAZAXATCCBSCZAZATCCBTCSCABTCPCYAUCRCRCTCABZATCYAYABBZABBSCRCQCXAQCR',
        message: 'DXCADSCACVAPBWCPAVABDWAUBRCEDACNBZANBWCABCDZCZCTCWBCCKBFDBCBDRBDCKDEDVCPBTBICDCQCWCXBLDEDNDTCYBQBNDJDUCNBUBQCPAHCACRC9DTAICWATBRCVBEDRBWCFCVCPCUBTBEDSCLBYAGCCDTANBZBBCCCCCKDICLBZBADICSCTBBDYBYBDCDDACTCTBRBCCHCUCCDJDTBECMDVCXBNDADCCECWCHCSCWBUCZBUAVBRBTAACBBNB9CVCSBWBED9DIDLBNBYCUBPAZBKBUCMDNDZAGDNBPCYCXBVCHDGCUAGCPAUAABMBBBQCDDSBWBBDUAFCQB9C9BACZCGCWBKBFCRBBBACZCQCPAXBTACDADADYC9CBBGCRCUADCCBZCPAADNBFC9CKBZCLDVAICHCCBAC9BXCRCXCGDICNDYCNBHCZCOBEDZCUCWAJDQCUBPAYCPBADDDRBKBYCXAPC9DUAPADDZATAZBCCMBPAICZAZCPCWCZAUBKBMDUCGCMBHCWCYBXBPCNDTCTBEDKBHCWAZAMDJDYBYCQCTAJDYCWARCABFDQCLDYAZBUCSC9DMD9CXATBBDWANBACZBNBWATCXAPB9DYCPBJDADXBADGCTBECUCNDTAYBUCFCGDVAADBCFCLDABSBRC9BPCYCBDLBBDDCYBZBIDPAYBOBIDEDXADDICPBQBICWCTAWCNDNBRB9BPAJDLBTBECADCBRCECFCABRBUCKBICZARCOBZANBECBBCBWBZBHCUAMDWAPCFCWCJDDCWCLBICDCUBPASBHDABHCPCXAZAFCEDDCCBPBKBZBCCOBVBIDCCXCBCFCMDSBLBIDZCOBMDSCBBMDYAKDQBXBZCSBXBSCVAQBVCVAYAPBACECNDQCLBZBJDCDPCMBWCIDLBZCNDDDWBLDDDKDFDXBACPCHDYBABSCXBADPANBVBUBWCHCFDLDOBACBDYCTBHDWAVBNBLBBDKBUCGCQCYCPCTBZCKBDCCBZCMBECADQBZCECTATAUCLBPCWAZAXBADPAHDDCACCDXAECGCKBRBCBKBXBYATCSBECFDMDZBLD9BGCFCCBFCVAICYCHCWACCNDSBYADDCCCDECEDNDECTAVCMDYARCACUATBFDFDABCBQCSCGCWCXCBCPAMBWCICTCFDKDSCRBNBVCACUBBDBCZBPBADJDZACCHCIDYCUABBHDJDMDRCECBDXAXBSBCBWAYBKBFCDCUBZCVBHCBCDDWCYAKBRBKDUAEDTBACABDDZAVBYBSCECDCEDBBICRCECRCXBACRBECUATBBBICLDGD9CRCUBYABBOBXBVCPBKDPCWCBBTCKBLBKDHDHCQBBCGCNBSBXBGCYCABKDJDXCVCJDCDVBYBOBSCLBPAUC9BYBAC9BZBABVADDPCPBSBNB9DUCZAFCYAZBUBHCUATAFDGCYBWBCCFC9DKDNBJDCCVBRCLDCDICPCBBNBHDKBADWADDOBKBMBRBZCFCWBLDNDABOBYBSCTCYCYBCDBDRBUCVARCCDCDVCFDQCHDACFDFCYCMDEDZCOBDC9BPCQCBDOBVBMBXBPCVCVCWCYBXCEDLBICTCDDUBLBWAGDWCABPCECLDZA9CVCTAGCXBFDRCPCQCVAFDCCXA9BVCWCLBCDADADADHDYAMBGCPBCBLBWACBHDCDIDRCYBGDTBCBZCWAXAADWCTC9BKDECNDCBLDPCECGCPCBBBDYCVCQBIDXCCDKB9BABTBVCCBWCFCCBBCWBVAPAEDXBJDGDWAVCDDSCWBYBZADCEDNDZCYCIDFCHCKBOBHCXCKDOBCBWCKBOBRBZCYBFDPBVCUACBCCOBLBTBBC9BJDMDBCTAPATBADZAGCMBPBSBXCTBJDHDNDGCBCLBOB9BND9DADXATBLBRCXBXBDDSBWBBCCDYBJDCBNDYBSBTCZBPBGC9CMDMBRCCD9BXC9BACPCPADDEDXAHCTCNBWCUAHCRCGDVBWCTBBDZBXAPBWABDRCSCFD9DWAUAXBUALBSCTBYC9DUALDZCCBQBJDUCXCZAQBTBRBJDBDGCDDICGDBCNBCDLBMBFDKDACMDDCKDMDKDDDUAVCVCYBDDCBACWCBDHCBDBBKDXCIDZBADWBOBQBQCND9BECYAPAEDVA9BGCLDICNDCDVASBPC'
    };

    let txInDB = {
        transaction_id: '',
        address: '',
        message: ''
    };

    Transactions.findOne({
        where: {
            transaction_id: txFromWebNode.transaction_id
        }
    }).then(result => {
        txInDB.transaction_id = result.get('transaction_id');
        txInDB.address = result.get('address');
        txInDB.message = result.get('message');
        check_tangle_for_work(txInDB);
    });

    return;
};

function check_tangle_for_work(txInDB) {

    let searchValues = {
        addresses: [txInDB.address]
    };

    console.log(searchValues);

    iota.api.findTransactionObjects(searchValues, function(result, result2) {
        console.log('in findTransactionObjects callback');
        console.log(result);
        console.log(result2);
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
