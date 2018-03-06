'use strict';
module.exports = (sequelize, DataTypes) => {
    var Transactions = sequelize.define('Transactions', {
        transaction_id: {
            type: DataTypes.INTEGER,
            primaryKey: true
        },
        need_requested: DataTypes.STRING,
        address: DataTypes.STRING,
        message: DataTypes.STRING,
        transaction_status: DataTypes.ENUM,
    }, {timestamps: true});
    Transactions.associate = function (models) {
        // associations can be defined here
    };
    return Transactions;
};