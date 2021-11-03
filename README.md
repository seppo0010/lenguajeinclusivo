# Juscaba

Juscaba permite obtener un expediente público y disponibilizarlo de una forma
más amigable y con un buscador propio.

## Por qué

Uno de los principios de la Justicia es su transparencia. No alcanza con tirar
archivos en un lugar. Los mismos deben ser usables para cumplir con este
requisito.

## Cómo funciona

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
