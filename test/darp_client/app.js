const express = require('express');
const bodyParser = require('body-parser');
const router = express.Router();
const app = express();
const port = 3000;

app.use(bodyParser.json());
app.use('/', router);

router.route('/service').post((req, resp) => {
    let adRequest = req.body;
    console.log(adRequest.request.object.metadata.name);
    console.log("/service");
    if (adRequest.request.object.metadata.labels.type === "good-service") {
        resp.send({'isAllowed': true, 'message': 'service has good-services label, allowing to proceed'});
    } else {
        resp.send({'isAllowed': false, 'message': 'service doesnt has good-services label, denying to proceed'});
    }
});

router.route('/service/check-labels').post((req, resp) => {
    let adRequest = req.body;
    console.log(adRequest.request.object.metadata.name);
    console.log("/labels");
    resp.send({'isAllowed': true, 'message': 'broken check labels test'});
});

router.route('/service/check-ports').post((req, resp) => {
    let adRequest = req.body;
    console.log("/ports");
    console.log(adRequest.request.object.metadata.name);
    resp.send({'isAllowed': true, 'message': 'all good check port test'});
});

app.listen(port, () => console.log(`listening on port ${port}!`));