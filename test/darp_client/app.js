const express = require('express');
const bodyParser = require('body-parser');
const router = express.Router();
const app = express();
const port = 3000;

app.use(bodyParser.json());
app.use('/', router);

router.route('/service').post((req, resp) => {
    console.log(JSON.stringify(req.body, null, 4));
    let adRequest = req.body;
    if (adRequest.request.object.metadata.labels.type === "good-service") {
        resp.send({'isAllowed': true, 'message': 'service has good-services label, allowing to proceed'});
    } else {
        resp.send({'isAllowed': false, 'message': 'service doesnt has good-services label, denying to proceed'});
    }
});

router.route('/service/check-labels').post((req, resp) => {
    resp.send({'isAllowed': false, 'message': 'broken'});
});

router.route('/service/check-ports').post((req, resp) => {
    resp.send({'isAllowed': true, 'message': 'all good'});
});

app.listen(port, () => console.log(`listening on port ${port}!`));