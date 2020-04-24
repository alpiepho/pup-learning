require("dotenv").config();
base = require('./base');

PUP_URL_BASE="https://www.linkedin.com/learning";
PUP_URL_LOGIN=PUP_URL_BASE+"/me";
PUP_URL_LOGOUT=PUP_URL_BASE+"/logout";
PUP_URL_COMPLETED=PUP_URL_BASE+"/me/completed";


// in ms
PAGE_WAIT = 1000;
PAGE_WAIT_LOGIN_DONE = 3000;
PAGE_WAIT_COMPLETED = 2000;


const process_login = async (browser, options) => {
  var waitMs = PAGE_WAIT + base.random_int(100);
  //console.log('process_login')
  const page = await base.browser_get(browser, PUP_URL_LOGIN, (waitMs));
  await base.process_options(browser, options);
  await page.type("#auth-id-input", process.env.PUP_USERNAME);
  await base.delay(waitMs);
  await page.click("#auth-id-button"); // Email page "Continue"
  await base.delay(waitMs);
  await base.process_options(browser, options);
  await page.type("#password", process.env.PUP_PASSWORD);
  await base.delay(waitMs);
  await page.click(".btn__primary--large"); // Password page "Continue"
  await base.delay(PAGE_WAIT_LOGIN_DONE);
  await base.process_options(browser, options);
  //console.log("process_login done")
};

const process_logout = async (browser, options) => {
  //console.log('process_logout')
  const page = await base.browser_get(browser, PUP_URL_LOGOUT, PAGE_WAIT);
  //console.log("process_logout done")
};

async function auto_scroll(page){
  await page.evaluate(async () => {
      await new Promise((resolve, reject) => {
          var totalHeight = 0;
          var distance = 400;
          var timer = setInterval(() => {
              var scrollHeight = document.body.scrollHeight;
              window.scrollBy(0, distance);
              totalHeight += distance;

              if(totalHeight >= scrollHeight){
                  clearInterval(timer);
                  resolve();
              }
          }, 500);
      });
  });
}

const process_completed = async (browser, options, data) => {
  //console.log("process_completed");
  const page = await base.browser_get(browser, PUP_URL_COMPLETED, PAGE_WAIT_COMPLETED);
  if (options.scrollToBottom) {
    await auto_scroll(page);
  }
  await base.delay(PAGE_WAIT_COMPLETED);
  await base.process_options(browser, options);

  const newdata = await page.evaluate(() => {
    let result = {};

    // parse: 'Learning History (108)'
    let count = document.querySelector('#ember160').innerText;
    result['count'] = count.replace(')','').split('(')[1];

    // parse: table of completed courses
    //  - course title
    //  - course author
    //  - course date recorded
    //  - course durration
    //  - completed date
    //  - course link
    // below requires navigation
    //  - course details
    //  - author LinkedIn link
    //  - course toc
    //    - sections
    //      - title
    //      - subsections
    //        - title
    //        - description
    //        - durration
    //  - course exercise files?
    //  - **could** also grab transcript???

    // course lings
    result['links'] = [...document.querySelectorAll('.lls-card-detail-card-body__headline a.card-entity-link')].map(elem => elem.href);

    return result;
  });
  // TODO: cleaner way to copy
  data['count'] = newdata['count'];
  // assemble nested data from lists, assume collated
  data['links'] = newdata['links'];

  // const links = await page.evaluate(
  //   () => [...document.querySelectorAll('h2 a')].map(elem => elem.href)
  // );
  // links.forEach(item => data['teams'].push(item));

  //console.log("process_completed done");
};

exports.process_login = process_login;
exports.process_logout = process_logout;
exports.process_completed = process_completed;
