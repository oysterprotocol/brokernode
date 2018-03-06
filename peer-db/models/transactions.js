'use strict';
const Sequelize = require('sequelize');

module.exports = (sequelize, DataTypes, Sequelize) => {
    var Transactions = sequelize.define('Transactions', {
        transaction_id: {
            type: Sequelize.UUID,
            defaultValue: Sequelize.UUIDV4,
            unique: true,
            primaryKey: true,
        },
        need_requested: DataTypes.STRING,
        address: DataTypes.STRING,
        message: Sequelize.TEXT('medium'),
        transaction_status: DataTypes.ENUM,
    }, {timestamps: true});
    Transactions.associate = function (models) {
        // associations can be defined here
    };
    return Transactions;
};