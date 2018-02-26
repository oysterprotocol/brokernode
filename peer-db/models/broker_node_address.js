'use strict';
module.exports = (sequelize, DataTypes) => {
    var Broker_node_address = sequelize.define('Broker_node_address', {
        id: {
            type: DataTypes.INTEGER,
            primaryKey: true
        },
        address: DataTypes.STRING
    }, {});
    Broker_node_address.associate = function (models) {
        // associations can be defined here
    };
    return Broker_node_address;
};