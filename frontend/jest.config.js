export default {
  testEnvironment: "jsdom",
  transform: {},
  moduleNameMapper: {
    "\\.(css|less)$": "<rootDir>/tests/__mocks__/styleMock.js"
  }
};
