import {proxy} from 'utils/queryUrlGenerator';
import AJAX from 'utils/ajax';
import _ from 'lodash';

export function getCpuAndLoadForHosts(proxyLink) {
  return proxy({
    source: proxyLink,
    query: `select mean(usage_user) from cpu where cpu = 'cpu-total' and time > now() - 10m group by host; select mean("load1") from "telegraf".."system" where time > now() - 10m group by host`,
    db: 'telegraf',
  }).then((resp) => {
    const hosts = {};
    const precision = 100;
    const cpuSeries = _.get(resp, ['data', 'results', '0', 'series'], []);
    const loadSeries = _.get(resp, ['data', 'results', '1', 'series'], []);
    cpuSeries.forEach((s) => {
      const meanIndex = s.columns.findIndex((col) => col === 'mean');
      hosts[s.tags.host] = {
        name: s.tags.host,
        cpu: (Math.round(s.values[0][meanIndex] * precision) / precision),
      };
    });

    loadSeries.forEach((s) => {
      const meanIndex = s.columns.findIndex((col) => col === 'mean');
      hosts[s.tags.host].load = (Math.round(s.values[0][meanIndex] * precision) / precision);
    });

    return hosts;
  });
}

export function getMappings() {
  return AJAX({
    method: 'GET',
    url: `/chronograf/v1/mappings`,
  });
}

export function getAppsForHosts(proxyLink, hosts, supportedApps) {
  const measurements = supportedApps.map((m) => `${m}$`).join('|');
  return proxy({
    source: proxyLink,
    query: `show series from /${measurements}/`,
    db: 'telegraf',
  }).then((resp) => {
    const newHosts = Object.assign({}, hosts);
    const allSeries = _.get(resp, ['data', 'results', '0', 'series', '0', 'values'], []);
    allSeries.forEach(([series]) => {
      const matches = series.match(/(\w*),.*,host=([^,]*)/);
      if (!matches || matches.length !== 3) { // eslint-disable-line no-magic-numbers
        return;
      }
      const app = matches[1];
      const host = matches[2];

      if (!newHosts[host]) {
        return;
      }
      if (!newHosts[host].apps) {
        newHosts[host].apps = [];
      }
      newHosts[host].apps = _.uniq(newHosts[host].apps.concat(app));
    });

    return newHosts;
  });
}
