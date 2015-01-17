'use strict';

/* https://github.com/angular/protractor/blob/master/docs/toc.md */

describe('my app', function() {

  browser.get('index.html');

  it('should automatically redirect to /register when location hash/fragment is empty', function() {
    expect(browser.getLocationAbsUrl()).toMatch("/register");
  });


  describe('register', function() {

    beforeEach(function() {
      browser.get('index.html#/register');
    });


    it('should render the registration form when user navigates to /register', function() {
      expect(element.all(by.css('[ng-view] h2')).first().getText()).
        toMatch('Group pre-registration');
    });

  });
});
