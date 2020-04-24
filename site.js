require("dotenv").config();
base = require('./base');

PUP_URL_BASE="https://www.linkedin.com/learning";
PUP_URL_LOGIN=PUP_URL_BASE+"/me";
PUP_URL_LOGOUT=PUP_URL_BASE+"/logout";

PUP_URL_COMPLETED=PUP_URL_BASE+"/me/completed";

PUP_URL = process.env.PUP_LEARNING_URL || "https://www.linkedin.com/learning/me";
PUP_URL_HREF = PUP_URL;
// PUP_URL_HREF = PUP_URL.replace('http:', 'https:').replace('www.', '')
PUP_URL_LOGIN = PUP_URL;


// in ms
PAGE_WAIT = 0;
//PAGE_WAIT_TAB = 2000;
PAGE_WAIT_LOGIN = 3000;



const process_login = async (browser, options) => {
  console.log('process_login')
  const page = await base.browser_get_simple(browser, PUP_URL_LOGIN, 2000);
  await page.type("#auth-id-input", process.env.PUP_USERNAME);
  await base.delay(1000);
  await page.click("#auth-id-button");
  await base.delay(2000);
  await page.type("#password", process.env.PUP_PASSWORD);
  await base.delay(1000);
  await page.click(".btn__primary--large");
  await base.delay(5000);

  
  //await page.waitForSelector('button[button_id="configure_view"]');
  //await base.process_options(browser, options);
  //console.log("process_login done")
};

const process_logout = async (browser, options) => {
  console.log('process_logout')
  const page = await base.browser_get_simple(browser, PUP_URL_LOGOUT, PAGE_WAIT);
  console.log("process_logout done")
};

async function autoScroll(page){
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

const process_completed = async (browser, data) => {
  console.log("process_completed");
  const page = await base.browser_get_simple(browser, PUP_URL_COMPLETED, 2000);
  //await page.waitForSelector('a[tab_name="stats"]');
  //await page.click('a[tab_name="stats"]');
  //await autoScroll(page);
  await base.delay(2000);

  const newdata = await page.evaluate(() => {
    let result = {};
    // like 'Learning History (108)'
    let temp = document.querySelector('#ember160').innerText;
    result['count'] = temp.replace(')','').split('(')[1];

    return result;
  });
  data['count'] = newdata['count'];

  // const links = await page.evaluate(
  //   () => [...document.querySelectorAll('h2 a')].map(elem => elem.href)
  // );
  // links.forEach(item => data['teams'].push(item));

  console.log("process_completed done");
};

exports.process_login = process_login;
exports.process_logout = process_logout;
exports.process_completed = process_completed;
