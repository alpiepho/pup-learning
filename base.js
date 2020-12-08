fsUtils = require("nodejs-fs-utils");
puppeteerC = require("puppeteer");
puppeteerF = require("puppeteer-firefox");



function random_int(max) {
  return Math.floor(Math.random() * Math.floor(max));
}

function delay(time) {
  return new Promise(function(resolve) {
    setTimeout(resolve, time);
  });
}

const browser_init = async (options) => {
  //console.log('browser_init')
  if (options.browserType == "chrome") {
    const browser = await puppeteerC.launch({
      headless: options.headless,
      ignoreHTTPSErrors: true,
      defaultViewport: null,
      args: [
        "--ignore-certificate-errors",
        "--ignore-certificate-errors-spki-list"
      ]
    });
    return browser;
  }
  if (options.browserType == "firefox") {
    const browser = await puppeteerF.launch({
      headless: options.headless,
      ignoreHTTPSErrors: true,
      defaultViewport: null,
      args: [
        "--ignore-certificate-errors",
        "--ignore-certificate-errors-spki-list"
      ]
    });
    return browser;
  }

  return null;
};

const browser_get = async (browser, href, waitTime) => {
  let page;
  try {
    console.log("browser_get " + href);
    page = (await browser.pages())[0];
    await page.goto(href);
    await delay(waitTime);
  } catch (err) {}
  return page;
};


const browser_close = async browser => {
  //console.log('browser_close')
  await browser.close();
};

const process_options = async (browser, options) => {
  //console.log('process_options')
  //console.log('process_options done')
};

exports.random_int = random_int;
exports.delay = delay;
exports.browser_init = browser_init;
exports.browser_get = browser_get;
exports.browser_close = browser_close;
exports.process_options = process_options;
