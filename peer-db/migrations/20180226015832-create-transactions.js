'use strict';
module.exports = {
    up: (queryInterface, Sequelize) => {
        return queryInterface.createTable('Transactions', {
            transaction_id: {
                allowNull: false,
                autoIncrement: true,
                primaryKey: true,
                type: Sequelize.INTEGER
            },
            need_requested: {
                type: Sequelize.STRING
            },
            work: {
                type: Sequelize.STRING
            },
            transaction_status: {
                type: Sequelize.ENUM,
                values: ['WAITING_FOR_ITEM_SELECTION',
                    'ITEM_SELECTED_PAYMENT_PENDING',
                    'ITEM_SELECTED_PAYMENT_REPORTED',
                    'ITEM_SELECTED_PAYMENT_CONFIRMED',
                    'ITEM_SENT',
                    'TRANSACTION_COMPLETE'],
                defaultValue: 'WAITING_FOR_ITEM_SELECTION'
            },
            createdAt: {
                allowNull: false,
                type: Sequelize.DATE
            },
            updatedAt: {
                allowNull: false,
                type: Sequelize.DATE
            }
        });
    },
    down: (queryInterface, Sequelize) => {
        return queryInterface.dropTable('Transactions');
    }
};

