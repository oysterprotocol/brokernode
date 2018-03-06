'use strict';
module.exports = {
    up: (queryInterface, Sequelize) => {
        return queryInterface.createTable('Transactions', {
            transaction_id: {
                allowNull: false,
                primaryKey: true,
                type: Sequelize.UUID,
                unique: true
            },
            need_requested: {
                type: Sequelize.STRING
            },
            address: {
                type: Sequelize.STRING
            },
            message: {
                type: Sequelize.TEXT('medium')
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

            // Timestamps
            createdAt: {
                allowNull: true,
                type: Sequelize.DATE
            },
            updatedAt: {
                allowNull: true,
                type: Sequelize.DATE
            }
        });
    },
    down: (queryInterface, Sequelize) => {
        return queryInterface.dropTable('Transactions');
    }
};

