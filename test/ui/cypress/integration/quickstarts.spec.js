import OpenshiftConsole from "../code/openshift/openshiftConsole";

describe('OCP UI for Serverless', () => {

  beforeEach(() => {
    const openShiftConsole = new OpenshiftConsole()
    openShiftConsole.login()
  })

  it.skip('has Serverless quickstarts', () => {
    cy.visit('/quickstart')
    cy.contains('Exploring Serverless applications')
  })
})
