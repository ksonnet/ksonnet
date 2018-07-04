function(moduleName, params)
    local prefix = if (moduleName == "/" || moduleName == "") then "" else "%s." % moduleName;

    local baseObject = if std.objectHas(params, "global")
        then {global: params.global}
        else {};

    baseObject + {
        components: {
            ["%s%s" % [prefix, key]]: params.components[key]
            for key in std.objectFieldsAll(params.components)
        },
    }