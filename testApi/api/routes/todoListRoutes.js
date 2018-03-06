'use strict';

module.exports = function (app) {


    var exchangeController = require('../controllers/todoListController');

    app.route('/givePeerId')
        .post(exchangeController.add_peer_id);
    app.route('/startTransaction')
        .post(exchangeController.start_transaction);
    app.route('/selectNeed')
        .post(exchangeController.need_selected);
    app.route('/confirmWork')
        .post(exchangeController.confirm_work);
};
