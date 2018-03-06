'use strict';

module.exports = (sequelize, DataTypes) => {
    var Peer_id = sequelize.define('PeerIds', {
        id: {
            type: DataTypes.UUID,
            defaultValue: DataTypes.UUIDV4,
            unique: true,
            primaryKey: true,
        },
        peer_id: {
            type: DataTypes.STRING,
            unique: true
        },
    }, {});
    Peer_id.associate = function (models) {
        // associations can be defined here
    };
    return Peer_id;
};