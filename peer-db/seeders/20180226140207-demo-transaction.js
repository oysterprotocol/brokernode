'use strict';

module.exports = {
    up: (queryInterface, Sequelize) => {
        return queryInterface.bulkInsert('Transactions', [{
            transaction_id: 6,
            need_requested: 'genesis_hash',
            work: 'TODO',
            transaction_status: 'WAITING_FOR_ITEM_SELECTION',
            createdAt: new Date(),
            updatedAt: new Date()
        }], {});
    },

    down: (queryInterface, Sequelize) => {
        return queryInterface.bulkDelete('Transactions', null, {});
    }
};
