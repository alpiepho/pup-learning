fs = require('fs');

base = require('./base');
site = require('./site');

const main = async () => {
  // INTERNAL OPTION
  options = { 
    browserType: "firefox", // "chrome, firefox"
    headless: false,
    screenshot: true, 
    screenshotDir: "/tmp/pup_learning_screenshots",
    scrollToBottom: false
  }
  const browser = await base.browser_init(options);
  options.version = await browser.version();
  console.log("options:");
  console.log(options);

  // login, get list of completed courses, logout
  data = {}
  await site.process_login(browser, options);
  await site.process_completed(browser, options, data);
  await site.process_logout(browser, options);
  await base.browser_close(browser);

  // TODO: generate artifacts from data
  // TODO: generate markdown (.mdx) for blog
  // TODO: generate html for deploy on GH Pages
  console.log("data:");
  console.log(JSON.stringify(data, null, space=2));

  console.log("done.");
};

main();

  
  