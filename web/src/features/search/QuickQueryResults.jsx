// QuickQueryResults reuses FindResults — quick query responses use the same
// PaginatedSearchResponse shape. This component is kept as a thin re-export
// so the file exists per the design doc, but ResultPanel routes quick_query
// results directly to FindResults.
export { default } from './FindResults';
