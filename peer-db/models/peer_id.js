'use strict';
module.exports = (sequelize, DataTypes) => {
  var Peer_id = sequelize.define('Peer_id', {
    id: {
        type: DataTypes.INTEGER,
        primaryKey: true
    },
    peer_id: DataTypes.STRING
  }, {});
  Peer_id.associate = function(models) {
    // associations can be defined here
  };
  return Peer_id;
};