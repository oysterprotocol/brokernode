'use strict';
module.exports = (sequelize, DataTypes) => {
  var Transactions = sequelize.define('Transactions', {
    transaction_id: {
        type: DataTypes.INTEGER,
        primaryKey: true
    },
    need_requested: DataTypes.STRING,
    work: DataTypes.STRING,
    transaction_status: DataTypes.ENUM
  }, {});
  Transactions.associate = function(models) {
    // associations can be defined here
  };
  return Transactions;
};