fsUtils = require("nodejs-fs-utils");
lighthouse = require('lighthouse');

puppeteerC = require("puppeteer");
puppeteerF = require("puppeteer-firefox");

// PUP_URL = process.env.PUP_URL;
// PUP_URL_HREF = PUP_URL;
// // PUP_URL_HREF = PUP_URL.replace('http:', 'https:').replace('www.', '')
// PUP_URL_LOGIN = PUP_URL;

// in ms
PAGE_WAIT = 0;
PAGE_WAIT_TAB = 2000;
PAGE_WAIT_LOGIN = 4000;

screenshot_count = 0;
lighthouse_count = 0;

function delay(time) {
  return new Promise(function(resolve) {
    setTimeout(resolve, time);
  });
}

const browser_init = async (options) => {
  //console.log('browser_init')

  if (options.screenshot) {
    try {
      fsUtils.removeSync(options.screenshotDir);
    } catch (err) {}
    try {
      fsUtils.mkdirsSync(options.screenshotDir);  
    } catch (err) {}
  }

  if (options.lighthouse) {
    try {
      fsUtils.removeSync(options.lighthouseDir);
    } catch (err) {}
    try {
      fsUtils.mkdirsSync(options.lighthouseDir);  
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

const browser_get_simple = async (browser, href, waitTime) => {
  let page;
  try {
    console.log("browser_get_simple " + href);
    //const page = await browser.newPage();
    page = (await browser.pages())[0];
    await page.goto(href);
    await delay(waitTime);
  } catch (err) {}
  return page;
};

const browser_get = async (browser, href, waitTime) => {
  let page;
  for (let i = 0; i < 3; i++) {
    try {
      console.log("browser_get " + href);
      //const page = await browser.newPage();
      page = (await browser.pages())[0];
      await page.goto(href);
      await delay(waitTime);
    } catch (err) {
      if (waitTime == 0) waitTime = 1000;
      else waitTime = waitTime * 2;
    }
  }
  return page;
};

const browser_close = async browser => {
  //console.log('browser_close')
  await browser.close();
};

const process_screenshot = async (browser, filename) => {
  //console.log('process_screenshot')
  const page = (await browser.pages())[0];
  await page.screenshot({ path: filename })
  //console.log('process_screenshot done')
};

const process_lighthouse = async (browser, filename) => {
  console.log('process_lighthouse')
  if (options.version.includes('Chrome')) {
    const page = (await browser.pages())[0];
    const url = page.url();
    const {lhr} = await lighthouse(url, {
      port: (new URL(browser.wsEndpoint())).port,
      output: 'json',
      logLevel: 'error',
    });
    //console.log(lhr);  
    console.log(`Lighthouse scores: ${Object.values(lhr.categories).map(c => c.score).join(', ')}`);
    fs.writeFileSync(filename, JSON.stringify(lhr), 'utf8')

    await browser_get(browser, url, PAGE_WAIT);
  }
  else {
    console.log("NOTE: Lighhouse only works with Chrome");
  }
  //console.log('process_lighthouse done')
};

const process_options = async (browser, options) => {
  //console.log('process_options')
  if (options.screenshot) {
    filename = options.screenshotDir + '/img' + screenshot_count + '.jpeg'
    await process_screenshot(browser, filename)
    screenshot_count += 1;
  }  
  if (options.lighthouse) {
    filename = options.lighthouseDir + '/lhr' + lighthouse_count + '.json'
    await process_lighthouse(browser, filename);
    lighthouse_count += 1;
  }  
  //console.log('process_options done')
};

exports.delay = delay;
exports.browser_init = browser_init;
exports.browser_get_simple = browser_get_simple;
exports.browser_get = browser_get;
exports.browser_close = browser_close;
exports.process_options = process_options;
