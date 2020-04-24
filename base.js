fsUtils = require("nodejs-fs-utils");

puppeteerC = require("puppeteer");
puppeteerF = require("puppeteer-firefox");

screenshot_count = 0;

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
  if (options.useSampleData) {
    return null;
  }
  if (options.screenshot) {
    try {
      fsUtils.removeSync(options.screenshotDir);
    } catch (err) {}
    try {
      fsUtils.mkdirsSync(options.screenshotDir);  
    } catch (err) {}
  }

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
    //const page = await browser.newPage();
    page = (await browser.pages())[0];
    await page.goto(href);
    await delay(waitTime);
  } catch (err) {}
  return page;
};

// const browser_get_retries = async (browser, href, waitTime, retries) => {
//   let page;
//   for (let i = 0; i < retries; i++) {
//     try {
//       console.log("browser_get " + href);
//       //const page = await browser.newPage();
//       page = (await browser.pages())[0];
//       await page.goto(href);
//       await delay(waitTime);
//     } catch (err) {
//       if (waitTime == 0) waitTime = 1000;
//       else waitTime = waitTime * 2;
//     }
//   }
//   return page;
// };

const browser_close = async browser => {
  //console.log('browser_close')
  if (options.useSampleData) {
    return null;
  }
  await browser.close();
};

const process_screenshot = async (browser, filename) => {
  //console.log('process_screenshot')
  const page = (await browser.pages())[0];
  await page.screenshot({ path: filename })
  //console.log('process_screenshot done')
};

const process_options = async (browser, options) => {
  //console.log('process_options')
  if (options.screenshot) {
    filename = options.screenshotDir + '/img' + screenshot_count + '.jpeg'
    await process_screenshot(browser, filename)
    screenshot_count += 1;
  }  
  //console.log('process_options done')
};

exports.random_int = random_int;
exports.delay = delay;
exports.browser_init = browser_init;
exports.browser_get = browser_get;
//exports.browser_get_retries = browser_get_retries;
exports.browser_close = browser_close;
exports.process_options = process_options;
