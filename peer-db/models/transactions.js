'use strict';

module.exports = (sequelize, DataTypes) => {
    var Transactions = sequelize.define('Transactions', {
        transaction_id: {
            type: DataTypes.UUID,
            defaultValue: DataTypes.UUIDV4,
            unique: true,
            primaryKey: true,
        },
        need_requested: DataTypes.STRING,
        address: DataTypes.STRING,
        message: DataTypes.TEXT('medium'),
        transaction_status: {
            type: DataTypes.ENUM,
            values: ['WAITING_FOR_ITEM_SELECTION',
                'ITEM_SELECTED_PAYMENT_PENDING',
                'ITEM_SELECTED_PAYMENT_REPORTED',
                'ITEM_SELECTED_PAYMENT_CONFIRMED',
                'ITEM_SENT',
                'TRANSACTION_COMPLETE'],
            defaultValue: 'WAITING_FOR_ITEM_SELECTION'
        },
    }, {timestamps: true});
    Transactions.associate = function (models) {
        // associations can be defined here
    };
    return Transactions;
};