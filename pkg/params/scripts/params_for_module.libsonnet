function(moduleName, input)
   local isModule(key) = std.length(std.split(key, ".")) > 1;

   local localizeKey(key) =
      if isModule(key)
      then
         local parts = std.split(key, ".");
         parts[std.length(parts)-1]
      else key;

   local findInRoot(key, value) =
      if isModule(key)
      then {[key]:null}
      else {[key]:value};

   local findInModule(moduleName, key, value) =
      if std.startsWith(key, moduleName)
      then {[localizeKey(key)]: value}
      else {[localizeKey(key)]: null};

   local findValue(moduleName, key, value) =
      if moduleName == "/"
      then findInRoot(key, value)
      else findInModule(moduleName, key, value);

   local fn(moduleName, params) = [
      findValue(moduleName, key, params.components[key])
      for key in std.objectFields(params.components)
   ];

   local foldFn(aggregate, object) =
      local o = {
         components+: {
            [x]:+ object[x]
            for x in std.objectFields(object)
            if object[x] != null
         },
      };

      aggregate + o;

   local a = fn(moduleName, input);

   local init = {components: {}};

   std.foldl(foldFn, a, init)

