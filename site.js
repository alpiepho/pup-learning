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
  await base.delay(12000);

  
  //await page.waitForSelector('button[button_id="configure_view"]');
  //await base.process_options(browser, options);
  //console.log("process_login done")
};

const process_logout = async (browser, options) => {
  console.log('process_logout')
  const page = await base.browser_get_simple(browser, PUP_URL_LOGOUT, PAGE_WAIT);
  console.log("process_logout done")
};

// const process_throbber = async page => {
//   //console.log("process_throbber");
//   // have to wait for "throbber" to show up and then be removed
//   try {
//     await page.waitForSelector(
//       'div[class="throbber_overlay"]',
//       (timeout = PAGE_WAIT_TAB)
//     );
//     console.log
//     await page.waitFor(
//       () =>
//         !document.querySelector(
//           'div[class="throbber_overlay"]')
//     );  
//   } catch (error) {}
//   //console.log("process_throbber done");
// };

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

const process_completed = async (browser, options) => {
  console.log("process_completed");
  const page = await base.browser_get_simple(browser, PUP_URL_COMPLETED, 2000);
  //await page.waitForSelector('a[tab_name="stats"]');
  //await page.click('a[tab_name="stats"]');
  await autoScroll(page);
  await base.delay(12000);
  console.log("process_completed done");
};

exports.process_login = process_login;
exports.process_logout = process_logout;
exports.process_completed = process_completed;
