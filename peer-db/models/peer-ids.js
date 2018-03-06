'use strict';
const Sequelize = require('sequelize');

module.exports = (sequelize, DataTypes, Sequelize) => {
    var Peer_id = sequelize.define('PeerIds', {
        id: {
            type: Sequelize.UUID,
            defaultValue: Sequelize.UUIDV4,
            unique: true,
            primaryKey: true,
        },
        peer_id: {
            type: DataTypes.STRING,
            unique: true
        },
    }, {});
    Peer_id.associate = function (models) {
        // associations can be defined here
    };
    return Peer_id;
};