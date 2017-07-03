FAQ

Q1: ¿Cómo y cuando se actualiza el número total de invitados (num_guests) y de asistentes (num_attendees)?

R1: Ambos son valores precalculados bien cuando se crea o modifica un evento, o cuando se lee de la base de datos. El número total de invitados se actualiza la primera vez que se crea el evento. Posteriormente, si se modifica el evento, el número total de invitados se actualizará también. En cuanto a num_attendees, inicialmente es 0 al crear el evento. Si se modifica un evento que ya tenía participantes confirmados el evento modificado tendrá el valor correcto. En ambos casos, tanto num_guests como num_attendees será correctos en el momento de leerse el evento de la base de datos.

--

Q2: ¿Qué pasa si se añaden nuevos invitados a un evento existente desde dos lugares a la vez?

R2: Desde el punto de vista de la base de datos no habría problemas de inconsistencia. Cada participante ocupa su propia fila dentro de la partición por lo que, siempre y cuando no se sobrescriba su respuesta o estado, no habría problemas en añadir usuarios al evento desde más de un lugar a la vez (ni siquiera cuando algunos usuarios coincidan).

Desde el punto de vista del objeto en memoria la cosa cambia. En este caso, si dos usuarios A y B invitan a sus amigos a un evento al mismo tiempo, asumiendo 3 invitados iniciales, que los nuevos invitados son distintos y que A invita a 3 usuarios y B invita a 4 usuarios, inicialmente al leer el evento de la base de datos, el objeto en memoria será consistente para ambos usuarios. Sin embargo tras modificarse el evento, el objeto en memoria del usuario A tendrá 6 invitados, mientras que el usuario B tendrá 7 invitados. En este caso, cuando el servidor envia el evento al usuario A y usuario B, sus bases de datos podrían ser inconsistentes.

Dado que el problema reside en que el objeto en memoria es diferente y que este puede llegar a transmitirse por la red a otros clientes, la solución pasaría aquí por enviar únicamente la información que ha cambiado y incluir un número de versión. De este modo, el usuario A únicamente enviaría que ha añadido a 2 usuarios y el usuario B únicamente enviaría que ha añadido a 3 usuarios. Cuando ambos mensajes lleguen a todos los participantes del evento, el resultado será que hay 5 usuarios nuevos, además de los 3 que ya tenían. Por tanto, los clientes y la base de datos del servidor estarán sincronizados.

--

Q3: ¿Qué pasa si se modifica el mismo evento desde dos sitios a la vez?

R3: La información del evento sería la de la última escritura en producirse. Si dos usuarios A y B modifican un evento al mismo tiempo, al enviarse la información al resto de participantes del evento algunos de estos llegarían a tener un estado inconsistente dependiendo del orden de llegada de los mensajes. Si por ejemplo la base de datos si hubiera quedado en el estado puesto por B, si un cliente recibe primero el mensaje del usuario B y luego el de A, la base de datos del cliente sería inconsistente.

Al igual que en el caso anterior, la solución pasa por controlar y versionar la información que se transmite al cliente para que este pueda decidir qué cambios aplica y cuales no. En este caso, todos los campos que componen la información del evento (description, start date, end date, event state y position) se escribirán en bloque. De este modo, el evento no podrá contener alguna información modificada por el usuario A y otra por el usuario B. Es decir, si el cliente A modifica la descripción y el cliente B modifica la fecha y la hora, nunca se realizará la mezcla de ambos cambios.

Otra solución a este problema es hacer que la escritura que llegue primero se haga efectiva, mientras que la segunda escritura reciba un error avisando de que la copia que intenta modificar ya no existe y por tanto debe volver a leer el evento.

--

Q4: ¿Qué pasa si se modifica un evento mientras un participante cambia su respuesta?

R4: La información que se puede modificar de un participante no solapa con la información que se puede modificar del evento. En concreto, durante la modificación del evento solo se pueden añadir nuevos participantes que se traduce en añadir una nueva fila (única para cada participante) a la lista de participantes del evento en la base de datos. Mientras que el cambio de respuesta o estado de un participante modificaría, dentro de una fila existente, la columna correspondiente. Si la misma fila y columna (celda) se modifican a la vez, la última escritura gana. En resumen, se pueden añadir nuevas participantes y modificar la respuesta o estado de un participante existente al mismo tiempo y sin se produzca ninguna inconsistencia en la base de datos.

No obstante, el objeto en memoria del evento modificado si que sería inconsistente dado que, si la respuesta de un participante cambió después de haberse leído el evento, el evento en memoria contendrá información sin actualizar. De nuevo, esto hay que tenerlo en cuenta si se decide transmitir al cliente el evento en memoria pues, al ser inconsistente, trasladaría dicha inconsistencia a la base de datos de los clientes. En este caso, si el cliente recibe primero la modificación del evento y luego el cambio en la respuesta, la información en el cliente será consistente con la del servidor. Si se recibe primero el cambio en la respuesta y luego la modificación ¡¡la información sería inconsistente!!

Como ya se ha comentado, una solución a este problema sería versionar los cambios de estado de cada participante. De este modo, el cliente podrá decidir ignorar cambios anteriores si la versión actual es superior y solo aplicar cambios con un número de version mayor que la actual.

NOTA: Si la información de un mismo participante cambia desde dos sitios a la vez, la última operación de escritura será la que se mantenga. Esto no es un problema ya que solo un participante puede cambiar su respuesta en el evento.

--

Q5: Cuando cambio el estado o la respuesta de un participante ¿debo actualizar también el evento asociado si ha sido cargado previamente?

R5: En el momento en que la información de un evento se lee de la base de datos, esta es consistente. Si el participante de un evento cambia de estado, naturalmente un evento leído previamente sería inconsistente con la información en la base de datos. En el caso de que el evento se modifique esto no supone ningún inconveniente ya que ninguna modificación del evento puede cambiar el estado del participante. El único caso en que puede resultar en inconsistencias sería si el evento se envía al cliente ya que estaría transmitiendo información antigua. En este caso, si se versiona el estado del participante el cliente podrá aplicar únicamente la información más actual e ignorar la antigua. 

--

Q6: ¿Quién se encarga de leer los próximos eventos en RAM para poblar los próximos eventos de cada usuario y los eventos activos?

R6:

--

Q7: ¿Tiene sentido devolver copia del evento y no trabajar sobre el evento existente?

R7: La inmutabilidad es una propiedad deseable dentro de un sistema concurrente para evitar condiciones de carrera. Del mismo modo, al tratarse también de un sistema distribuido no tiene sentido que todos los nodos compartan el mismo espacio de eventos en memoria. Esto implica que los cambios de un evento en la memoria de un nodo no se sincronizan con el mismo evento en la memoria de otro nodo, ni siquiera dentro del mismo nodo. Es decir, en un mismo instante t dos nodos cualesquiera pueden tener en memoria una versión diferente del mismo evento. No obstante, los casos en los que esto puede suceder son limitados y no provocan efectos secundarios:

a) cassandra tiene consistencia 'eventually consistency'. Cuando las escrituras en un evento aún no se han propagado por todos los nodos una lectura posterior a la escritura aún puede recuperar una versión anterior de un evento;

b) cuando dos nodos leen la misma versión de un evento pero ambos la modifican. Esto provoca que cada nodo tenga un evento diferente en su memoria RAM. Sin embargo, no supone un problema dado que al escribir el evento en la base de datos la última escritura será la que prevalezca. NOTA: La escritura que fue inmediatamente sobrescrita no recibiría ningún error. Por lo que respecta al cliente se realizó correctamente. !!!! REVISAR !!!!

c) cuando un nodo lee un evento en RAM y otro nodo modifica y guarda el estado de un participante del mismo evento. Como la información del participante forma parte del evento, el evento en RAM del primer nodo será inconsistente con lo que hay realmente en la base de datos. Esto tampoco supone un problema porque si se envía el evento al cliente, el estado del participante estará versionado. 

--

Q8: ¿Qué pasa si se modifica un evento en RAM y se envía al cliente antes de guardarse en la base de datos?

R8: En este caso, cuando se modifica el evento su versión será indeterminada hasta que el evento se guarde en la base de datos. De este modo, si el evento se envía al cliente este lo ignoraría al no poder decidir si aplicar el cambio o descartarlo.

--

Q9: En el caso de los builders que modifican un objeto, ¿debe poder devolver un error en algún caso concreto o el responsable de todo es la llamada a Build?

R9: - La llamada a Build() es la que construye el objeto y la que debe comprobar cualquier otra condición, por tanto es la responsable de generar el error. No obstante, en algunos casos es interesante poder devolver también un error antes de llegar siquiera a obtener un constructor. Por ejemplo, si el evento ya ha empezado y, por tanto, no se debe poder modificar.

--

Q10: ¿Qué pasa si un usuario sube la misma imagen dos veces?

R10: Un mismo usuario no debería tener la misma imagen repetida en sus blobs

*/

/*

REQUISITOS

 - Todos los objetos que puedan leerse de la base de datos deben tener un flag (loaded) que indique si la instancia
 ha sido recuperada de la base de datos o no. Esto determinará al guardar dicho objeto si se guarda como uno nuevo o debe modificarse el existente.

 - Cuando se modifica un objeto, debe añadirse una lista con el atributo que se ha modificado y su antiguo valor. Esto será útil en ocasiones donde haya que actualizar una instancia antigua con una recien modificada.
*/


/*
 * Example: Load an event
 */
storedEvent, err := model.EventManager.LoadEvent(event_id)


/*
 * Example: Load multiple events
 */
storedEvent, err := model.EventManager.LoadAllEvents(id1, id2, id2, ...)


/*
 * Example: Create a new event
 */
newEvent, err := model.EventManager.NewEvent() // Returns an EventBuilder
    .SetAuthor(author)
    .SetCreatedDate(cd) // Optional, current date by default
    .SetModificationDate(md) // Optional, created date by default
    .SetStartDate(sd)
    .SetEndDate(ed)
    .SetDescription(d)
    .AddParticipant(userID1)
    .AddParticipant(userID2)
    .SetCancelled(true) // Optional, false by default
    .Build() // Checks are performed here, trigger error earlier as possible

storedEvent, err := model.EventManager.SaveEvent(newEvent) // Checks are performed here again (maybe not all of them)


/*
 * Example: Modify an event
 */
storedEvent, err := model.EventManager.LoadEvent(event_id)
modEvent, err := model.EventManager.ModifyEvent(storedEvent) // Returns an EventBuilder (with other defaults)
    .SetStartDate(cd)
    .SetEndDate(ed)
    .SetModificationDate(md) // Optional, current date by default
    .SetDescription(d)
    .SetCancelled(true) // Optional, false by default
    .AddParticipant(userID3)
    .Build()

storedEvent, err := model.EventManager.SaveEvent(modEvent)


/*
 * Example: Attach image to an existing event
 */

// Create a new blob
blob, err := model.BlobManager.New(imageBytes)
err := model.BlobManager.Save(blob)

// or load an existing one
blob, err := model.BlobManager.Load(blobID) 

// Attach it to an existing event
storedEvent, err := model.EventManager.LoadEvent(event_id)
modEvent, err := model.EventManager.ModifyEvent(storedEvent)
    .AttachImage(blob, model.MainImage)
    .Build() // Triggers an error if blob is not saved in DB before attach to event
storedEvent, err := model.EventManager.Save(modEvent) // Triggers an error if blob is not saved in DB before attach to event


/*
 * Example: Read image of an existing event
 */
storedEvent, err := model.EventManager.LoadEvent(event_id)
blobID := storedEvent.Blob(model.MainImage)
blob, err := model.BlobManager.Load(blobID) // blob.Data() contains binary data


/*
 * Example: Read participants
 */
storedEvent, err := model.EventManager.LoadEvent(event_id)
for _, p := storedEvent.Participants() {
    fmt.Println(p.Id(), p.Name(), p.Response(), p.Status())
}


/*
 * Example: Change participant
 */
p, err := model.EventManager.LoadParticipant(user_id, event_id)
newP, err := model.EventManager.ModifyParticipant(p) // Returns ParticipantBuilder
    .SetResponse(...)
    .SetStatus(...)
    .Build()

storedP, err := model.EventManager.SaveParticipant(newP) // TODO: Should also change any event previously loaded?


/*
 * Example: Read recent events for a user
 */
events, err := model.EventManager.LoadRecentEventsForUser(userID)


/*
 * Example: Read events history for a user
 */
option := model.ReadBackward // or model.ReadForward
events, err := model.EventManager.LoadEventsHistoryForUser(userID, fromDate, option)
