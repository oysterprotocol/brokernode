'use strict';
const Sequelize = require('sequelize');

module.exports = (sequelize, DataTypes, Sequelize) => {
    var Broker_node_address = sequelize.define('BrokerNodeAddresses', {
        id: {
            type: Sequelize.UUID,
            defaultValue: Sequelize.UUIDV4,
            unique: true,
            primaryKey: true,
        },
        address: {
            type: DataTypes.STRING,
            unique: true
        },
    }, {});
    Broker_node_address.associate = function (models) {
        // associations can be defined here
    };
    return Broker_node_address;
};