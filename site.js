require("dotenv").config();
axios = require('axios');
base = require('./base');
cheerio = require('cheerio');
fs = require('fs');

PUP_URL_BASE="https://www.linkedin.com/learning";
PUP_URL_LOGIN=PUP_URL_BASE+"/me";
PUP_URL_LOGOUT=PUP_URL_BASE+"/logout";
PUP_URL_COMPLETED=PUP_URL_BASE+"/me/completed?trk=nav_neptune_learning";


// in ms
PAGE_WAIT = 4000;
PAGE_WAIT_LOGIN = 4000;
PAGE_WAIT_LOGIN_DONE = 4000;
PAGE_WAIT_COMPLETED = 4000;

const process_manual_login = async (browser, options) => {
  var waitMs = PAGE_WAIT_LOGIN + base.random_int(1000);
  //console.log('process_manual_login')
  await base.browser_get(browser, PUP_URL_LOGIN, (waitMs));

  console.log("FINISH LOGIN");
  await base.delay(60000);

  await base.delay(PAGE_WAIT_LOGIN_DONE);
  await base.process_options(browser, options);
  //console.log("process_manual_login done")
};

const process_logout = async (browser, options) => {
  //console.log('process_logout')
  const page = await base.browser_get(browser, PUP_URL_LOGOUT, PAGE_WAIT);
  //console.log("process_logout done")
};

async function auto_scroll(page) {
  await page.evaluate(async () => {
    await new Promise((resolve, reject) => {
      var totalHeight = 0;
      var distance = 400;
      var timer = setInterval(() => {
        var scrollHeight = document.body.scrollHeight;
        window.scrollBy(0, distance);
        totalHeight += distance;

        if (totalHeight >= scrollHeight) {
          clearInterval(timer);
          resolve();
        }
      }, 1000);
    });
  });
}

const process_course_details = async (options, href) => {
  console.log("process_course_details: " + href);
  var newdata = {};
  newdata['linkedin'] = "";
  newdata['details'] = "";
  
  newdata = await axios.get(href).then((response) => {
    var results = {};
    results['linkedin'] = "";
    results['details'] = "";

    // parse course details:
    // TODO:
    //  - course toc
    //    - sections
    //      - title
    //      - subsections
    //        - title
     //        - durration
    //  - course exercise files?
    //  - **could** also grab transcript???
    const $ = cheerio.load(response.data);
    elements = $('p.section-container__description');
    innerText = $(elements[0]).text()
    results['details'] = innerText.trim();

    try {
      elements = $('li.course-instructors__list-item a');
      results['linkedin'] = $(elements[0]).attr('href');
    } catch {};
    return results;
  })
  .catch(error => { console.log(error); return {}})

  //console.log("process_course_details done");
  return [newdata['linkedin'], newdata['details']];
};

const save_thumb = async (url, path) => {
  const writer = fs.createWriteStream(path)
  const response = await axios({
    url,
    method: 'GET',
    responseType: 'stream'
  })
  response.data.pipe(writer)
};

const process_completed = async (browser, options, data) => {
  //console.log("process_completed");
  var newdata;

  const page = await base.browser_get(
    browser,
    PUP_URL_COMPLETED,
    PAGE_WAIT_COMPLETED
  );

  if (options.scrollToBottom) {
    await auto_scroll(page);
  }

  try {
    // HACK: should get a count (clicking on a class might click all, not sure)
    await page.click('.lls-card-child-content__button'); // Show content
  } catch {}

  await base.delay(PAGE_WAIT_COMPLETED);
  await base.process_options(browser, options);

  newdata = await page.evaluate(() => {
    let result = {};

    result['completed-courses'] = []
    let card_conts = document.querySelectorAll('.lls-card-detail-card');
    for (i=0; i<card_conts.length; i++) {
      let entry = {};
      entry['title'] = '';
      entry['link'] = '';
      entry['author'] = '';
      entry['released-date'] = '';
      entry['duration'] = '';
      entry['completed-date'] = '';
      entry['img'] = '';
      entry['linkedin'] = '';
      entry['details'] = '';
      entry['title'] = card_conts[i].querySelector('.lls-card-headline').innerText;
      temp = card_conts[i].querySelector('a.entity-link').href;
      entry['link'] = temp;
      temp = card_conts[i].querySelector('.lls-card-authors');
      if (temp) entry['author'] = temp.innerText.replace('By: ','');
      temp = card_conts[i].querySelector('.lls-card-released-on');
      if (temp) entry['released-date'] = temp.innerText.replace('Released ','');
      temp = card_conts[i].querySelector('.lls-card-duration-label');
      if (temp) entry['duration'] = temp.innerText;
      temp = card_conts[i].querySelector('.lls-card-completion-state--completed');
      if (temp) entry['completed-date'] = temp.innerText.replace('Completed ','');
      temp = card_conts[i].querySelector('img');
      if (temp) entry['img'] = temp.src;
      result['completed-courses'].push(entry);
    }
    return result;
  });
  //console.log(newdata);

  if (options.gatherThumbs) {
    for (i=0; i<newdata['completed-courses'].length; i++) {
      entry = newdata['completed-courses'][i];
      if (entry['img']) {
        entry['img_guid'] = entry['img'].split('/')[5];
        entry['img_file'] = './images/' + entry['img_guid'] + '.jpg';
        path = './public/images/' + entry['img_guid'] + '.jpg';
        await save_thumb(entry['img'], path);
      }
    }
  }    

  if (options.gatherDetails) {
    for (i=0; i<newdata['completed-courses'].length; i++) {
      if (!newdata['completed-courses'][i]['details']) {
        console.log(i);
        await base.delay(2000);
        [temp1, temp2] = await process_course_details(options, newdata['completed-courses'][i]['link']);
        newdata['completed-courses'][i]['linkedin'] = temp1;
        newdata['completed-courses'][i]['details'] = temp2;
      }
    }
  }


  data['completed-courses'] = newdata['completed-courses'];
  //console.log("process_completed done");
};



exports.process_manual_login = process_manual_login;
exports.process_logout = process_logout;
exports.process_course_details = process_course_details;
exports.process_completed = process_completed;
