'use strict';

module.exports = (sequelize, DataTypes) => {
    var Broker_node_address = sequelize.define('BrokerNodeAddresses', {
        id: {
            type: DataTypes.UUID,
            defaultValue: DataTypes.UUIDV4,
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