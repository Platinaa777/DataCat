export enum DataSourceDriver {
  PROMETHEUS = 'prometheus',
  ELASTIC_SEARCH = 'elasticsearch',
  JAEGER = 'jaeger',
}

export interface DataSource {
  id: string;
  name: string;
  driver: DataSourceDriver;
  connectionUrl: string;
}
