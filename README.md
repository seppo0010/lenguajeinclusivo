# Juscaba

JUSCABA nace del deseo de hacer mas accesible la justicia argentina. Es un
robot que extrae toda la informacion publica de una serie de expedientes
judiciales de la justicia de la ciudad de buenos aires y los presenta de una
forma mas amigable tanto para usuarios que para investigadores.

## Por qué

Uno de los principios de la Justicia es su transparencia. No alcanza con tirar
archivos en un lugar. Los mismos deben ser usables para cumplir con este
requisito.

Si le interesa un expediente que no esta cargado (la mayoria) escribanos
[aquí](https://github.com/odia/juscaba/issues/new) o baje este proyecto y
ejecutelo en sus propios servidores. Más información sobre cómo hacerlo más
abajo.

## Como se usa ?

Seleccione el expediente que le interesa y podra buscar texto en todos sus
actuaciones y documentos asociados. tambien podra bajarse las fuentes
originales de esos documentos (PDFs).

## Cómo funciona (tecnicamente)

El procedimiento consta de tres pasos:

* El primer paso es obtener los documentos. Para ello se busca la ficha y las
actuaciones. De ella se buscan todos los archivos adjuntos.

* Luego se crea un índice para minisearch.

* Por último se crea una aplicación web para verlo.

## ¡Quiero mi expediente!

El procedimiento para obtener tu propio expediente debe ser algo asî.

```
EXPEDIENTE=123456-7890-1
EXPEDIENTE_NOMBRE="Mi expediente"
docker build -t juscaba .
docker run -v $(pwd):/tmp/juscaba juscaba ./run.sh $EXPEDIENTE
sudo chown -R $(whoami):$(whoami) web/build
echo "[{\"id\":\"${EXPEDIENTE_NOMBRE\",\"file\":\"$EXPEDIENTE\"}]" > web/build/data/expedientes.json
```

Y ahí en `web` estaría la página estática con toda la información disponible.`
