'use strict';
module.exports = (sequelize, DataTypes) => {
    var Transactions = sequelize.define('Transactions', {
        transaction_id: {
            type: DataTypes.INTEGER,
            primaryKey: true
        },
        need_requested: DataTypes.STRING,
        work: DataTypes.STRING,
        transaction_status: DataTypes.ENUM,
        createdAt: {
            type: DataTypes.DATE,
            allowNull: true,
            defaultValue: sequelize.literal('NOW()')
        },
        updatedAt: {
            type: DataTypes.DATE,
            allowNull: true,
            defaultValue: sequelize.literal('NOW()')
        }
    }, {
        timestamps: true
    });
    Transactions.associate = function (models) {
        // associations can be defined here
    };
    return Transactions;
};