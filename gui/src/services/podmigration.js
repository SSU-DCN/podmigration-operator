import RestClient from "./restClient";

// export default class Podmigration {
export default class Podmigration {
  constructor(options) {
    const opts = options || {};

    this.options = opts;
    this.client = new RestClient(opts);
  }

  buildUrl(path) {
    return this.client.buildUrl(path);
  }

  /**
   * Info
   */
  getInfo() {
    return this.client.get("/");
  }

  getStatus() {
    return this.client.status("/");
  }

  /**
   * Podmigrations
   */
  // get a list of all Podmigrations
  getAllPodmigrations(params) {
    return this.client.get("/api/Podmigrations", params);
  }
  createPodmigration(params) {
    console.log(params);
    return this.client.post("/api/Podmigrations", params);
  }
}
