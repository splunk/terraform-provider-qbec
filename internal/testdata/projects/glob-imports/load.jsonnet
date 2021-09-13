local parseYaml = std.native('parseYaml');
local strMap = import 'glob-importstr:*.yaml';

std.foldl(function(prev, key) prev { [key]: parseYaml(strMap[key])[0] }, std.objectFields(strMap), {})
